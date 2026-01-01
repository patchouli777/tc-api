package livestream

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// redis set to map username to livestream id
type userToIdStore struct {
	rdb *redis.Client
}

func (r *userToIdStore) Delete(ctx context.Context, username string) *redis.IntCmd {
	return r.rdb.Del(ctx, r.Key(username))
}

func (r *userToIdStore) TxDelete(ctx context.Context, tx redis.Pipeliner, username string) *redis.IntCmd {
	return tx.Del(ctx, r.Key(username))
}

func (r *userToIdStore) Key(username string) string {
	return fmt.Sprintf("livestreams_ids:%s", username)
}

func (r *userToIdStore) Add(ctx context.Context, username string, lsId string) (*redis.StatusCmd, error) {
	return r.rdb.Set(ctx, r.Key(username), lsId, 0), nil
}

func (r *userToIdStore) TxAdd(ctx context.Context, tx redis.Pipeliner, username string, lsId string) *redis.StatusCmd {
	return tx.Set(ctx, r.Key(username), lsId, 0)
}

func (r *userToIdStore) Get(ctx context.Context, username string) (*string, error) {
	res, err := r.rdb.Get(ctx, r.Key(username)).Result()
	if err != nil {
		return nil, err
	}

	return &res, nil
}
