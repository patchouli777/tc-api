package livestream

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// hashset to store livestreams by their id
//
// key is "livestreams:<id>"
type livestreamStore struct {
	rdb *redis.Client
}

func (r *livestreamStore) addTx(ctx context.Context, tx redis.Pipeliner, ls Livestream) *redis.IntCmd {
	return tx.HSet(ctx, r.key(strconv.Itoa(int(ls.Id))), ls)
}

func (r *livestreamStore) list(ctx context.Context, ids []string) ([]Livestream, error) {
	cmds, err := r.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, id := range ids {
			pipe.HGetAll(ctx, r.key(id))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	livestreams := make([]Livestream, len(ids))
	for i, cmd := range cmds {
		var ls Livestream
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

func (r *livestreamStore) get(ctx context.Context, lsId string) (*Livestream, error) {
	var ls Livestream

	err := r.rdb.HGetAll(ctx, r.key(lsId)).Scan(&ls)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errNotFound
		}

		return nil, err
	}

	return &ls, nil
}

func (r *livestreamStore) updateViewers(ctx context.Context, lsId string, viewers int) error {
	return r.rdb.HSet(ctx, r.key(lsId), "viewers", viewers).Err()
}

func (r *livestreamStore) updateFieldTx(ctx context.Context, tx redis.Pipeliner, id string, values map[string]any) error {
	return tx.HSet(ctx, r.key(id), values).Err()
}

func (r *livestreamStore) updateThumbnail(ctx context.Context, lsId, thumbnail string) error {
	return r.rdb.HSet(ctx, r.key(lsId), "thumbnail", thumbnail).Err()
}

func (r *livestreamStore) key(id string) string {
	return fmt.Sprintf("livestreams:%s", id)
}

func (r *livestreamStore) deleteTx(ctx context.Context, tx redis.Pipeliner, lsId string) *redis.IntCmd {
	return tx.Del(ctx, r.key(lsId))
}
