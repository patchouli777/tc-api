package livestream

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// livestream ids set
type idStore struct {
	rdb *redis.Client
}

func (r *idStore) key() string {
	return "livestream_ids"
}

func (r *idStore) addTx(ctx context.Context, tx redis.Pipeliner, id string) *redis.IntCmd {
	return tx.SAdd(ctx, r.key(), id)
}

func (r *idStore) getAll(ctx context.Context) ([]string, error) {
	res, err := r.rdb.SMembers(ctx, r.key()).Result()
	return res, err
}

func (r *idStore) deleteTx(ctx context.Context, tx redis.Pipeliner, id string) *redis.IntCmd {
	return tx.SRem(ctx, r.key(), id)
}
