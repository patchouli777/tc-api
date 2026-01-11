package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// hashset to store livestreams by their id
//
// key is "livestreams:<id>"
type livestreamStore struct {
	rdb *redis.Client
}

func (r *livestreamStore) add(ctx context.Context, ls Livestream) *redis.IntCmd {
	return r.rdb.HSet(ctx, r.key(strconv.Itoa(int(ls.Id))), ls)
}

func (r *livestreamStore) addTx(ctx context.Context, tx redis.Pipeliner, ls Livestream) *redis.IntCmd {
	return tx.HSet(ctx, r.key(strconv.Itoa(int(ls.Id))), ls)
}

func (r *livestreamStore) list(ctx context.Context, ids []string) ([]Livestream, error) {
	pipe := r.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))

	for i, id := range ids {
		cmds[i] = pipe.HGetAll(ctx, r.key(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	livestreams := make([]Livestream, len(ids))
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				slog.Error("inconsistent state: livestream id is present in sorted set, but livestream is not present in livestream store")
				continue
			}

			slog.Error("error executing cmd from pipeline", sl.Err(err))
			continue
		}

		ls, err := r.deserealize(val)
		if err != nil {
			return nil, err
		}
		livestreams[i] = *ls
	}

	return livestreams, nil
}

func (r *livestreamStore) listTx(ctx context.Context, tx redis.Pipeliner, ids []string) ([]Livestream, error) {
	cmds := make([]*redis.MapStringStringCmd, len(ids))

	for i, id := range ids {
		cmds[i] = tx.HGetAll(ctx, r.key(id))
	}

	_, err := tx.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	livestreams := make([]Livestream, len(ids))
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				slog.Error("inconsistent state: livestream id is present in sorted set, but livestream is not present in livestream store")
				continue
			}

			slog.Error("transaction failed", sl.Err(err))
			continue
		}

		ls, err := r.deserealize(val)
		if err != nil {
			return nil, err
		}
		livestreams[i] = *ls
	}

	return livestreams, nil
}

func (r *livestreamStore) get(ctx context.Context, lsId string) (*Livestream, error) {
	res, err := r.rdb.HGetAll(ctx, r.key(lsId)).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errNotFound
		}

		return nil, err
	}

	return r.deserealize(res)
}

func (r *livestreamStore) getTx(ctx context.Context, tx redis.Pipeliner, lsId string) (*Livestream, error) {
	res, err := tx.HGetAll(ctx, r.key(lsId)).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errNotFound
		}

		return nil, err
	}

	return r.deserealize(res)
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

func (r *livestreamStore) deserealize(re map[string]string) (*Livestream, error) {
	var c Category
	if err := json.Unmarshal([]byte(re["category"]), &c); err != nil {
		return nil, err
	}

	var u User
	if err := json.Unmarshal([]byte(re["user"]), &u); err != nil {
		return nil, err
	}

	var ls Livestream
	ls.Category = c
	ls.User = u

	id, err := strconv.Atoi(re["id"])
	if err != nil {
		return nil, err
	}

	startedAt, err := strconv.Atoi(re["started_at"])
	if err != nil {
		return nil, err
	}

	ls.Id = int32(id)
	ls.StartedAt = int64(startedAt)
	ls.Thumbnail = re["thumbnail"]
	ls.Title = re["title"]

	viewers, err := strconv.Atoi(re["viewers"])
	if err != nil {
		return nil, err
	}

	ls.Viewers = int32(viewers)

	return &ls, nil
}

func (r *livestreamStore) key(id string) string {
	return fmt.Sprintf("livestreams:%s", id)
}

func (r *livestreamStore) delete(ctx context.Context, lsId string) *redis.IntCmd {
	return r.rdb.Del(ctx, r.key(lsId))
}

func (r *livestreamStore) deleteTx(ctx context.Context, tx redis.Pipeliner, lsId string) *redis.IntCmd {
	return tx.Del(ctx, r.key(lsId))
}
