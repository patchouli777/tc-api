package test

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"
	application "twitchy-api/internal/app"
	streamservermock "twitchy-api/internal/external/streamserver/mock"
	"twitchy-api/internal/lib/setup"

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
	logger  *slog.Logger
	app     *application.App
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cfg := application.GetConfig()
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	dockerPool, err := initDockerPool()
	if err != nil {
		log.Fatalf("unable to init docker pool: %v", err)
	}
	pool = dockerPool

	redisRes, err := initRedis(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("unable to init redis: %v", err)
	}

	defer func() {
		if err := pool.Purge(redisRes); err != nil {
			log.Fatalf("unable to purge redis resource: %v", err)
		}
	}()

	pgRes, err := initPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("unable to init postgres: %v", err)
	}

	defer func() {
		if err := pool.Purge(pgRes); err != nil {
			log.Fatalf("unable to purge pg resource: %v", err)
		}
	}()

	go func() {
		err = streamservermock.Run(ctx, logger)
		if err != nil {
			log.Fatalf("unable to init stream server mock: %v", err)
		}
	}()

	app, err = application.NewApp(logger,
		rclient,
		pgpool,
		cfg)
	if err != nil {
		// TODO:
	}

	setup.RecreateSchema(pgpool, rclient)
	// setup.Populate(ctx,
	// 	pgpool,
	// 	app.AuthService,
	// 	app.StreamServerAdapter,
	// 	app.CategoryRepo,
	// 	app.FollowRepo,
	// 	app.UserRepo,
	// 	"http://127.0.0.1:1985/api/streams")

	// app.Init(ctx, cfg.Update)
	authMw := application.NewAuthMiddleware(logger, false)
	handler := app.CreateHandler(authMw)
	ts = httptest.NewServer(handler)
	defer ts.Close()

	m.Run()
}

func initDockerPool() (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// creates resource AND connects to redis. redis client is available as global variable "rclient"
func initRedis(ctx context.Context, cfg application.RedisConfig) (*dockertest.Resource, error) {
	redisRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, err
	}

	if err = pool.Retry(func() error {
		rclient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", redisRes.GetPort("6379/tcp")),
		})

		return rclient.Ping(ctx).Err()
	}); err != nil {
		return nil, err
	}

	return redisRes, nil
}

// creates resource AND connects to postgres. pgxpool is available as global variable "pgpool"
func initPostgres(ctx context.Context, cfg application.PostgresConfig) (*dockertest.Resource, error) {
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
		return nil, err
	}

	hostAndPort := pgRes.GetHostPort("5432/tcp")
	postgresConnURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		hostAndPort,
		cfg.Name)
	// databaseUrl := fmt.Sprintf("postgres://less:123@%s/twitchclone?sslmode=disable", hostAndPort)

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
		return nil, err
	}

	return pgRes, nil
}
