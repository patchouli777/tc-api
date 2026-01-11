package livestream

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// sorted set to store livestream ids sorted by viewers count
type sortedIDStore struct {
	rdb *redis.Client
}

func (r *sortedIDStore) get(ctx context.Context, category string, page, count int) ([]string, error) {
	ids, err := r.rdb.ZRevRangeByScore(ctx, r.key(category), &redis.ZRangeBy{
		Offset: int64((page - 1) * count),
		Count:  int64(count),
		Min:    "0",
		Max:    "999999999",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to find livestream in category %s: %v", category, err)
	}

	return ids, nil
}

func (r *sortedIDStore) getTx(ctx context.Context, tx redis.Pipeliner, category string, page, count int) *redis.StringSliceCmd {
	return tx.ZRevRangeByScore(ctx, r.key(category), &redis.ZRangeBy{
		Offset: int64((page - 1) * count),
		Count:  int64(count),
		Min:    "0",
		Max:    "999999999",
	})
}

func (r *sortedIDStore) add(ctx context.Context, categoryLink string, score int32, id string) *redis.IntCmd {
	return r.rdb.ZAdd(ctx, r.key(categoryLink), redis.Z{
		Score:  float64(score),
		Member: id,
	})
}

func (r *sortedIDStore) addTx(ctx context.Context, tx redis.Pipeliner, categoryLink string, score int32, id string) *redis.IntCmd {
	return tx.ZAdd(ctx, r.key(categoryLink), redis.Z{
		Score:  float64(score),
		Member: id,
	})
}

func (r *sortedIDStore) delete(ctx context.Context, categoryLink string) *redis.IntCmd {
	return r.rdb.Del(ctx, r.key(categoryLink))
}

func (r *sortedIDStore) deleteTx(ctx context.Context, tx redis.Pipeliner, categoryLink string) *redis.IntCmd {
	return tx.Del(ctx, r.key(categoryLink))
}

func (r *sortedIDStore) key(category string) string {
	return fmt.Sprintf("livestreams_ids_sorted:%s", category)
}
