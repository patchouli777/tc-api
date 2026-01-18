package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"main/internal/app"
	"main/internal/external/auth"
	"main/internal/lib/setup"
	"main/internal/lib/sl"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func main() {
	cfg := app.GetConfig()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		// https://github.com/redis/go-redis/issues/3536
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	postgresConnURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Name)

	pool, err := pgxpool.New(ctx, postgresConnURL)
	if err != nil {
		log.Error("unable to create connection pool", sl.Err(err))
		os.Exit(1)
	}

	setup.RecreateSchema(pool, rdb)

	app := app.NewApp(ctx,
		log,
		rdb,
		pool,
		&auth.ClientMock{},
		cfg.Env,
		cfg.InstanceID.String(),
		cfg.StreamServer,
		cfg.Asynq)

	streamServerBaseUrl := fmt.Sprintf("http://%s:%s%s", cfg.StreamServer.Host, cfg.StreamServer.Port, cfg.StreamServer.Endpoint)
	setup.Populate(ctx, pool,
		app.AuthService,
		app.StreamServerAdapter,
		app.CategoryRepo,
		app.FollowRepo,
		app.UserRepo,
		streamServerBaseUrl)

	os.Exit(0)
}
