package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// sorted set to store livestream ids sorted by viewers count
type sortedIDStore struct {
	rdb *redis.Client
}

func (r *sortedIDStore) get(ctx context.Context, category string, page, count int) ([]int, error) {
	res, err := r.rdb.ZRevRangeByScore(ctx, r.key(category), &redis.ZRangeBy{
		Offset: int64((page - 1) * count),
		Count:  int64(count),
		Min:    "0",
		Max:    "999999999",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to find livestream in category %s: %v", category, err)
	}

	ids := make([]int, len(res))
	for i, idStr := range res {
		id, _ := strconv.Atoi(idStr)
		ids[i] = id
	}

	return ids, nil
}

func (r *sortedIDStore) addTx(ctx context.Context, tx redis.Pipeliner, categoryLink string, score int, id int) *redis.IntCmd {
	return tx.ZAdd(ctx, r.key(categoryLink), redis.Z{
		Score:  float64(score),
		Member: id,
	})
}

func (r *sortedIDStore) deleteTx(ctx context.Context, tx redis.Pipeliner, categoryLink string) *redis.IntCmd {
	return tx.Del(ctx, r.key(categoryLink))
}

func (r *sortedIDStore) key(category string) string {
	return fmt.Sprintf("livestreams_ids_sorted:%s", category)
}

type sortedIDAllStore struct {
	rdb *redis.Client
}

func (r *sortedIDAllStore) get(ctx context.Context, page, count int) ([]int, error) {
	res, err := r.rdb.ZRevRangeByScore(ctx, r.key(), &redis.ZRangeBy{
		Offset: int64((page - 1) * count),
		Count:  int64(count),
		Min:    "0",
		Max:    "999999999",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to find livestreams at all: %v", err)
	}

	ids := make([]int, len(res))
	for i, idStr := range res {
		id, _ := strconv.Atoi(idStr)
		ids[i] = id
	}

	return ids, nil
}

func (r *sortedIDAllStore) addTx(ctx context.Context, tx redis.Pipeliner, score int, id int) *redis.IntCmd {
	return tx.ZAdd(ctx, r.key(), redis.Z{
		Score:  float64(score),
		Member: id,
	})
}

func (r *sortedIDAllStore) deleteTx(ctx context.Context, tx redis.Pipeliner) *redis.IntCmd {
	return tx.Del(ctx, r.key())
}

func (r *sortedIDAllStore) key() string {
	return fmt.Sprintf("livestreams_ids_sorted:all")
}
