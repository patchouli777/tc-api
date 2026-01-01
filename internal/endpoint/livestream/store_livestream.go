package livestream

import (
	"context"
	"encoding/json"
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

func (r *livestreamStore) Add(ctx context.Context, ls Livestream) *redis.IntCmd {
	return r.rdb.HSet(ctx, r.Key(strconv.Itoa(int(ls.Id))), ls)
}

func (r *livestreamStore) TxAdd(ctx context.Context, tx redis.Pipeliner, ls Livestream) *redis.IntCmd {
	return tx.HSet(ctx, r.Key(strconv.Itoa(int(ls.Id))), ls)
}

func (r *livestreamStore) List(ctx context.Context, ids []string) ([]Livestream, error) {
	pipe := r.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))

	for i, id := range ids {
		cmds[i] = pipe.HGetAll(ctx, r.Key(id))
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

func (r *livestreamStore) Get(ctx context.Context, lsId string) (*Livestream, error) {
	re, err := r.rdb.HGetAll(ctx, r.Key(lsId)).Result()
	if err != nil {
		return nil, err
	}

	return r.deserealize(re)
}

func (r *livestreamStore) UpdateViewers(ctx context.Context, lsId string, viewers int) error {
	return r.rdb.HSet(ctx, r.Key(lsId), "viewers", viewers).Err()
}

func (r *livestreamStore) UpdateThumbnail(ctx context.Context, lsId, thumbnail string) error {
	return r.rdb.HSet(ctx, r.Key(lsId), "thumbnail", thumbnail).Err()
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

func (r *livestreamStore) Key(id string) string {
	return fmt.Sprintf("livestreams:%s", id)
}

func (r *livestreamStore) Delete(ctx context.Context, lsId string) *redis.IntCmd {
	return r.rdb.Del(ctx, r.Key(lsId))
}

func (r *livestreamStore) TxDelete(ctx context.Context, tx redis.Pipeliner, lsId string) *redis.IntCmd {
	return tx.Del(ctx, r.Key(lsId))
}
