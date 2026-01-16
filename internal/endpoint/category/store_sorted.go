package category

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// store id of categories sorted by viewers
type sortedStore struct {
	rdb *redis.Client
}

func (s *sortedStore) addTx(ctx context.Context, tx redis.Pipeliner, viewers int, id string) *redis.IntCmd {
	return tx.ZAdd(ctx, s.key(), redis.Z{
		Score:  float64(viewers),
		Member: id,
	})
}

func (s *sortedStore) get(ctx context.Context, start int64, count int64) ([]string, error) {
	return s.rdb.ZRevRangeByScore(ctx, s.key(), &redis.ZRangeBy{
		Min:    "0",
		Max:    "99999999",
		Offset: start,
		Count:  count,
	}).Result()
}

func (s *sortedStore) updateViewers(ctx context.Context, id string, viewers int32) error {
	return s.rdb.ZAdd(ctx, s.key(), redis.Z{
		Score:  float64(viewers),
		Member: id,
	}).Err()
}

func (s *sortedStore) deleteTx(ctx context.Context, tx redis.Pipeliner, id string) *redis.IntCmd {
	return tx.ZRem(ctx, s.key(), id)
}

func (s *sortedStore) key() string {
	return "categories_links_sorted"
}
