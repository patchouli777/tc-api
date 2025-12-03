package main

import (
	"context"
	"log"
	"os"

	"main/internal/app"
	"main/internal/endpoint/auth"
	"main/internal/lib/setup"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := app.GetConfig()
	ctx := context.Background()

	// ----------------Постгрес и редис старт----------------
	rdb := redis.NewClient(cfg.Redis)
	pool, err := pgxpool.New(ctx, cfg.Postgres.ConnURL)
	if err != nil {
		log.Fatalf("unable to create pg pool: %v", err)
	}
	defer pool.Close()

	setup.RecreateSchema(pool, rdb)

	srvcs := app.InitServices(ctx, cfg.Log,
		cfg.InstanceID.String(),
		cfg.Update.LivestreamsTimer,
		&auth.GRPCClientMock{},
		rdb,
		pool)

	setup.Populate(ctx, pool,
		srvcs.Auth,
		srvcs.Livestream,
		srvcs.Category,
		srvcs.Follow,
		srvcs.User)

	os.Exit(0)
}
