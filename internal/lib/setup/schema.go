package setup

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func RecreateSchema(pool *pgxpool.Pool, rdb *redis.Client) {
	ctx := context.Background()

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.DownTo(db, "internal/external/db/migrations", 0); err != nil {
		panic(err)
	}

	if err := goose.UpTo(db, "internal/external/db/migrations", 99999); err != nil {
		panic(err)
	}

	rdb.FlushAll(ctx)
}
