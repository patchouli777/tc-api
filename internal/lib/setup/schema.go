package setup

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func RecreateSchema(pool *pgxpool.Pool, rdb *redis.Client) {
	ctx := context.Background()
	b, err := os.ReadFile(".\\internal\\external\\db\\scripts\\remove_all.sql")
	if err != nil {
		log.Printf("unable to read file deleting schema in internal: %v", err)

		// b, err = os.ReadFile(root + "\\remove_all.sql")
		b, err = os.ReadFile(".\\remove_all.sql")
		if err != nil {
			log.Fatalf("unable to read file deleting schema in current dir: %v", err)
		}
	}

	r, err := pool.Query(context.Background(), string(b))
	if err != nil {
		log.Fatalf("unable to delete schema: %v", err)
	}
	r.Close()
	log.Println("schema deleted")

	rdb.FlushAll(ctx)

	// b, err = os.ReadFile(root + "\\internal\\db\\scripts\\schema.sql")
	b, err = os.ReadFile(".\\internal\\external\\db\\scripts\\schema.sql")
	if err != nil {
		log.Printf("unable to read file creating schema in internal: %v", err)

		// b, err = os.ReadFile(root + "\\schema.sql")
		b, err = os.ReadFile(".\\schema.sql")
		if err != nil {
			log.Fatalf("unable to read file deleting schema in current dir: %v", err)
		}
	}
	_, err = pool.Exec(context.Background(), string(b))
	if err != nil {
		log.Fatalf("unable to create schema: %v", err)
	}
	log.Println("schema created")
}
