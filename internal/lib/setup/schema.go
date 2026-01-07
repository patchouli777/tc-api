package setup

import (
	"context"
	"log"
	"main/pkg/util"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func RecreateSchema(pool *pgxpool.Pool, rdb *redis.Client) {
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("unable to acquire connection: %v", err)
	}

	root := util.GetProjectRoot()

	b, err := os.ReadFile(root + "\\internal\\db\\scripts\\remove_all.sql")
	if err != nil {
		log.Printf("unable to read file deleting schema in internal: %v", err)

		b, err = os.ReadFile(root + "\\remove_all.sql")
		if err != nil {
			log.Fatalf("unable to read file deleting schema in current dir: %v", err)
		}
	}

	r, err := conn.Query(context.Background(), string(b))
	if err != nil {
		log.Fatalf("unable to delete schema: %v", err)
	}
	r.Close()
	log.Println("schema deleted")

	rdb.FlushAll(ctx)

	b, err = os.ReadFile(root + "\\internal\\db\\scripts\\schema.sql")
	if err != nil {
		log.Printf("unable to read file creating schema in internal: %v", err)

		b, err = os.ReadFile(root + "\\schema.sql")
		if err != nil {
			log.Fatalf("unable to read file deleting schema in current dir: %v", err)
		}
	}
	_, err = conn.Exec(context.Background(), string(b))
	if err != nil {
		log.Fatalf("unable to create schema: %v", err)
	}
	log.Println("schema created")
}
