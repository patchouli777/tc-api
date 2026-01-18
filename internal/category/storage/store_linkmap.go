package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// store map link-to-id
type linkMap struct {
	rdb *redis.Client
}

func (s *linkMap) get(ctx context.Context, link string) (string, error) {
	return s.rdb.Get(ctx, link).Result()
}

func (s *linkMap) addTx(ctx context.Context, tx redis.Pipeliner, link string, id int) *redis.StatusCmd {
	return tx.Set(ctx, link, id, 0)
}

func (s *linkMap) deleteTx(ctx context.Context, tx redis.Pipeliner, link string) *redis.IntCmd {
	return tx.Del(ctx, link)
}
