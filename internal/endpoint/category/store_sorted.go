package category

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// store id of categories sorted by viewers
type sortedStore struct {
	rdb *redis.Client
}

func (s *sortedStore) add(ctx context.Context, viewers int, id string) error {
	return s.rdb.ZAdd(ctx, s.key(), redis.Z{
		Score:  float64(viewers),
		Member: id,
	}).Err()
}

func (s *sortedStore) get(ctx context.Context, start int64, count int64) ([]string, error) {
	return s.rdb.ZRevRangeByScore(ctx, s.key(), &redis.ZRangeBy{
		Min:    "0",
		Max:    "99999999",
		Offset: start,
		Count:  count,
	}).Result()
}

func (s *sortedStore) delete(ctx context.Context) error {
	return errors.New("not implemented")
}

func (s *sortedStore) key() string {
	return "categories_links_sorted"
}
