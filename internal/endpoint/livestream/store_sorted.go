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

func (r *sortedIDStore) Key(category string) string {
	return fmt.Sprintf("livestreams_ids_sorted:%s", category)
}

func (r *sortedIDStore) Delete(ctx context.Context, categoryLink string) *redis.IntCmd {
	return r.rdb.Del(ctx, r.Key(categoryLink))
}

func (r *sortedIDStore) Get(ctx context.Context, category string, page, count int) ([]string, error) {
	ids, err := r.rdb.ZRangeByScore(ctx, r.Key(category), &redis.ZRangeBy{
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

func (r *sortedIDStore) Add(ctx context.Context, categoryLink string, score int32, id string) *redis.IntCmd {
	return r.rdb.ZAdd(ctx, r.Key(categoryLink), redis.Z{
		Score:  float64(score),
		Member: id,
	})
}
