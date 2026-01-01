package livestream

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// livestream ids set
type idStore struct {
	rdb *redis.Client
}

func (r *idStore) key() string {
	return "livestream_ids"
}

func (r *idStore) Add(ctx context.Context, id string) error {
	return r.rdb.SAdd(ctx, r.key(), id).Err()
}

func (r *idStore) TxAdd(ctx context.Context, tx redis.Pipeliner, id string) *redis.IntCmd {
	return tx.SAdd(ctx, r.key(), id)
}

func (r *idStore) GetAll(ctx context.Context) ([]string, error) {
	res, err := r.rdb.SMembers(ctx, r.key()).Result()
	return res, err
}

func (r *idStore) Exist(ctx context.Context, id string) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *idStore) Delete(ctx context.Context, id string) error {
	return r.rdb.SRem(ctx, r.key(), id).Err()
}

func (r *idStore) TxDelete(ctx context.Context, tx redis.Pipeliner, id string) *redis.IntCmd {
	return tx.SRem(ctx, r.key(), id)
}
