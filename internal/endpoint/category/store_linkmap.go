package category

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

func (s *linkMap) delete(ctx context.Context, link string) error {
	return s.rdb.Del(ctx, link).Err()
}

func (s *linkMap) add(ctx context.Context, link string, id int) error {
	return s.rdb.Set(ctx, link, id, 0).Err()
}
