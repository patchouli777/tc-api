package test

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/app"
	"main/internal/lib/setup"
	"main/internal/lib/sl"
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

var pgpool *pgxpool.Pool
var rclient *redis.Client
var ts *httptest.Server

func TestMain(m *testing.M) {
	var exitCode int

	defer func() { os.Exit(exitCode) }()

	ctx := context.Background()
	cfg := app.GetConfig()
	log := cfg.Log

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Error("unable to construct dockertest pool", sl.Err(err))
		return
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Error("unable to connect to docker", sl.Err(err))
		return
	}

	redisRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Error("unable to start redis resource", sl.Err(err))
		return
	}

	if err = pool.Retry(func() error {
		rclient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", redisRes.GetPort("6379/tcp")),
		})

		// rclient = redis.NewClient(cfg.Redis)

		return rclient.Ping(ctx).Err()
	}); err != nil {
		log.Error("unable to connect to docker", sl.Err(err))
		return
	}

	defer func() {
		if err = pool.Purge(redisRes); err != nil {
			log.Error("unable to purge redis resource", sl.Err(err))
			return
		}
	}()

	pgRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=123",
			"POSTGRES_USER=less",
			"POSTGRES_DB=twitchclone",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Error("unable to start pg resource", sl.Err(err))
		return
	}

	hostAndPort := pgRes.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://less:123@%s/twitchclone?sslmode=disable", hostAndPort)
	log.Info("connection to databse on url", slog.String("url", databaseUrl))

	// pgRes.Expire(30) // Tell docker to hard kill the container in 30 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 3000 * time.Second
	if err = pool.Retry(func() error {
		pool, err := pgxpool.New(ctx, databaseUrl)
		if err != nil {
			log.Error("unable to create connection pool", sl.Err(err))
			return err
		}

		conn, err := pool.Acquire(ctx)
		if err != nil {
			log.Error("unable to acquire connection", sl.Err(err))
			return err
		}
		defer conn.Release()

		pgpool = pool

		return nil
	}); err != nil {
		log.Error("unable to connect to docker", sl.Err(err))
		return
	}

	defer func() {
		if err := pool.Purge(pgRes); err != nil {
			log.Error("unable to purge pg resource", sl.Err(err))
			return
		}
	}()

	// API_URL = "localhost:" + cfg.HTTP.Port

	grpcClient, err := app.NewGRPClient(cfg.Log, cfg.Env, cfg.GRPC)
	if err != nil {
		log.Error("unable to purge pg resource", sl.Err(err))
		return
	}

	srvcs := app.InitServices(ctx, cfg.Log,
		cfg.InstanceID.String(),
		cfg.Update.LivestreamsTimer,
		grpcClient,
		rclient,
		pgpool)

	setup.RecreateSchema(pgpool, rclient)
	setup.Populate(ctx, pgpool,
		srvcs.Auth,
		srvcs.Livestream,
		srvcs.Category,
		srvcs.Follow,
		srvcs.User)

	handler := app.CreateHandler(ctx, cfg, srvcs)
	ts = httptest.NewServer(handler)
	defer ts.Close()

	exitCode = m.Run()
}
