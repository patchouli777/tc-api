package storage

import (
	"context"
	"errors"
	"fmt"
	d "twitchy-api/internal/livestream/domain"

	"github.com/redis/go-redis/v9"
)

// hashset to store livestreams by their id
//
// key is "livestreams:<id>"
type livestreamStore struct {
	rdb *redis.Client
}

func (r *livestreamStore) addTx(ctx context.Context, tx redis.Pipeliner, ls d.Livestream) *redis.IntCmd {
	return tx.HSet(ctx, r.key(ls.Id), ls)
}

func (r *livestreamStore) list(ctx context.Context, ids []int) ([]d.Livestream, error) {
	cmds, err := r.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, id := range ids {
			pipe.HGetAll(ctx, r.key(id))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	livestreams := make([]d.Livestream, len(ids))
	for i, cmd := range cmds {
		var ls d.Livestream
		cm, ok := cmd.(*redis.MapStringStringCmd)
		if !ok {
			return nil, errors.New("not ok")
		}

		err := cm.Scan(&ls)
		if err != nil {
			return nil, err
		}

		livestreams[i] = ls
	}

	return livestreams, nil
}

func (r *livestreamStore) get(ctx context.Context, lsId int) (*d.Livestream, error) {
	var ls d.Livestream

	err := r.rdb.HGetAll(ctx, r.key(lsId)).Scan(&ls)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, d.ErrNotFound
		}

		return nil, err
	}

	return &ls, nil
}

func (r *livestreamStore) exists(ctx context.Context, id int) (int64, error) {
	return r.rdb.Exists(ctx, r.key(id)).Result()
}

func (r *livestreamStore) updateViewers(ctx context.Context, id int, viewers int) error {
	return r.rdb.HSet(ctx, r.key(id), "viewers", viewers).Err()
}

func (r *livestreamStore) updateFieldTx(ctx context.Context, tx redis.Pipeliner, id int, values map[string]any) error {
	return tx.HSet(ctx, r.key(id), values).Err()
}

func (r *livestreamStore) updateThumbnail(ctx context.Context, id int, thumbnail string) error {
	return r.rdb.HSet(ctx, r.key(id), "thumbnail", thumbnail).Err()
}

func (r *livestreamStore) key(id int) string {
	return fmt.Sprintf("livestreams:%d", id)
}

func (r *livestreamStore) deleteTx(ctx context.Context, tx redis.Pipeliner, id int) *redis.IntCmd {
	return tx.Del(ctx, r.key(id))
}
