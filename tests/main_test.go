package test

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/app"
	"main/internal/endpoint/auth"
	"main/internal/lib/setup"
	"main/internal/lib/sl"
	"main/internal/lib/streamservermock"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
)

var (
	pgpool  *pgxpool.Pool
	rclient *redis.Client
	pool    *dockertest.Pool
	ts      *httptest.Server
	log     *slog.Logger
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cfg := app.GetConfig()

	log = app.NewLogger(cfg.Logger)
	log.With(slog.String("env", "tests"))
	slog.SetDefault(log)

	dockerPool, err := initDockerPool()
	if err != nil {
		log.Error("unable to init docker pool", sl.Err(err))
		os.Exit(1)
	}
	pool = dockerPool

	redisRes, err := initRedis(ctx, cfg.Redis)
	if err != nil {
		log.Error("unable to init redis", sl.Err(err))
		os.Exit(1)
	}

	defer func() {
		if err := pool.Purge(redisRes); err != nil {
			log.Error("unable to purge redis resource", sl.Err(err))
		}
	}()

	pgRes, err := initPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Error("unable to init postgres", sl.Err(err))
		os.Exit(1)
	}

	defer func() {
		if err := pool.Purge(pgRes); err != nil {
			log.Error("unable to purge pg resource", sl.Err(err))
		}
	}()

	grpcClient, err := initGRPC(cfg.Env, cfg.AuthServiceMock, cfg.GRPC)
	if err != nil {
		log.Error("unable to init grpc client", sl.Err(err))
		os.Exit(1)
	}

	go func() {
		err = streamservermock.Run(ctx)
		if err != nil {
			log.Error("unable to init stream server mock", sl.Err(err))
			os.Exit(1)
		}
	}()

	srvcs := app.InitApp(ctx, log,
		cfg.InstanceID.String(),
		cfg.Env,
		grpcClient,
		rclient,
		pgpool)

	setup.RecreateSchema(pgpool, rclient)
	// setup.Populate(ctx, pgpool,
	// 	srvcs.Auth,
	// 	srvcs.StreamServerAdapter,
	// 	srvcs.Category,
	// 	srvcs.Follow,
	// 	srvcs.User)
	// srvcs.StreamServerAdapter.Update(ctx, cfg.Update.LivestreamsTimeout)
	// srvcs.CategoryUpdater.Update(ctx, cfg.Update.CategoriesTimeout)

	authMw := app.NewAuthMiddleware(log, false)

	handler := app.CreateHandler(ctx, log, authMw, srvcs)
	ts = httptest.NewServer(handler)
	defer ts.Close()

	m.Run()
}

func initDockerPool() (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Error("unable to construct dockertest pool", sl.Err(err))
		return nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Error("unable to connect to docker", sl.Err(err))
		return nil, err
	}

	return pool, nil
}

// creates resource AND connects to redis. redis client is available as global variable "rclient"
func initRedis(ctx context.Context, cfg app.RedisConfig) (*dockertest.Resource, error) {
	redisRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Error("unable to start redis resource", sl.Err(err))
		return nil, err
	}

	if err = pool.Retry(func() error {
		rclient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", redisRes.GetPort("6379/tcp")),
		})

		return rclient.Ping(ctx).Err()
	}); err != nil {
		log.Error("unable to connect to docker", sl.Err(err))
		return nil, err
	}

	return redisRes, nil
}

// creates resource AND connects to postgres. pgxpool is available as global variable "pgpool"
func initPostgres(ctx context.Context, cfg app.PostgresConfig) (*dockertest.Resource, error) {
	pgRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			fmt.Sprintf("POSTGRES_PASSWORD=%s", cfg.Password),
			fmt.Sprintf("POSTGRES_USER=%s", cfg.User),
			fmt.Sprintf("POSTGRES_DB=%s", cfg.Name),
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Error("unable to start pg resource", sl.Err(err))
		return nil, err
	}

	hostAndPort := pgRes.GetHostPort("5432/tcp")
	postgresConnURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		hostAndPort,
		cfg.Name)
	// databaseUrl := fmt.Sprintf("postgres://less:123@%s/twitchclone?sslmode=disable", hostAndPort)
	log.Info("connection to databse on url", slog.String("url", postgresConnURL))

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 30 * time.Second
	if err = pool.Retry(func() error {
		pool, err := pgxpool.New(ctx, postgresConnURL)
		if err != nil {
			return err
		}

		conn, err := pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		pgpool = pool

		return nil
	}); err != nil {
		log.Error("unable to connect to docker", sl.Err(err))
		return nil, err
	}

	return pgRes, nil
}

func initGRPC(env string, mock bool, cfg app.GrpcClientConfig) (auth.Client, error) {
	return app.NewAuthClient(log, env, mock, cfg)
}
