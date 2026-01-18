package storage

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// redis set to map username to livestream id
type userToIdStore struct {
	rdb *redis.Client
}

func (r *userToIdStore) addTx(ctx context.Context, tx redis.Pipeliner, username string, lsId int) *redis.StatusCmd {
	return tx.Set(ctx, r.key(username), lsId, 0)
}

func (r *userToIdStore) get(ctx context.Context, username string) (int, error) {
	res, err := r.rdb.Get(ctx, r.key(username)).Int()
	if err != nil {
		return -1, err
	}

	return res, nil
}

func (r *userToIdStore) deleteTx(ctx context.Context, tx redis.Pipeliner, username string) *redis.IntCmd {
	return tx.Del(ctx, r.key(username))
}

func (r *userToIdStore) key(username string) string {
	return fmt.Sprintf("livestreams_ids:%s", username)
}
