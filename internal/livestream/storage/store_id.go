package storage

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// livestream ids set
type idStore struct {
	rdb *redis.Client
}

func (r *idStore) key() string {
	return "livestream_ids"
}

func (r *idStore) addTx(ctx context.Context, tx redis.Pipeliner, id int) *redis.IntCmd {
	return tx.SAdd(ctx, r.key(), id)
}

func (r *idStore) getAll(ctx context.Context) ([]int, error) {
	res, err := r.rdb.SMembers(ctx, r.key()).Result()
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(res))
	for i, idStr := range res {
		id, _ := strconv.Atoi(idStr)
		ids[i] = id
	}

	return ids, err
}

func (r *idStore) deleteTx(ctx context.Context, tx redis.Pipeliner, id int) *redis.IntCmd {
	return tx.SRem(ctx, r.key(), id)
}
