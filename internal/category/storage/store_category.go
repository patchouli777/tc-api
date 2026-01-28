package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	d "twitchy-api/internal/category/domain"

	"github.com/redis/go-redis/v9"
)

// store categories retrievable by id
type categoryStore struct {
	rdb *redis.Client
}

func (r *categoryStore) addTx(ctx context.Context, tx redis.Pipeliner, cat d.Category) *redis.IntCmd {
	idStr := strconv.Itoa(int(cat.Id))
	return tx.HSet(ctx, r.key(idStr), cat)
}

// TODO: is there any better way???
// redis marshaling doesn't work with slices and it is big bad
// cmd must be a result of HGetAll
func (r *categoryStore) parseCategory(cmd *redis.MapStringStringCmd) (*d.Category, error) {
	res, err := cmd.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, d.ErrNotFound
		}

		return nil, err
	}

	if len(res) == 0 {
		return nil, d.ErrNotFound
	}

	var category d.Category
	err = cmd.Scan(&category)
	if err != nil {
		return nil, err
	}

	if tagsJSON, ok := res["tags"]; ok {
		json.Unmarshal([]byte(tagsJSON), &category.Tags)
	}

	return &category, nil
}

func (r *categoryStore) get(ctx context.Context, id string) (*d.Category, error) {
	cmd := r.rdb.HGetAll(ctx, r.key(id))
	return r.parseCategory(cmd)
}

func (r *categoryStore) list(ctx context.Context, ids []string) ([]d.Category, error) {
	pipe := r.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))

	for i, id := range ids {
		cmds[i] = pipe.HGetAll(ctx, r.key(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	categories := make([]d.Category, len(ids))
	for i, cmd := range cmds {
		category, err := r.parseCategory(cmd)
		if err != nil {
			return nil, err
		}

		categories[i] = *category
	}

	return categories, nil
}

func (r *categoryStore) updateViewers(ctx context.Context, id string, viewers int) error {
	return r.rdb.HSet(ctx, r.key(id), "viewers", viewers).Err()
}

func (r *categoryStore) update(ctx context.Context, id string, values map[string]any) error {
	return r.rdb.HSet(ctx, r.key(id), values).Err()
}

func (r *categoryStore) deleteTx(ctx context.Context, tx redis.Pipeliner, id string) *redis.IntCmd {
	return tx.Del(ctx, r.key(id))
}

func (r *categoryStore) key(id string) string {
	return fmt.Sprintf("categories:%s", id)
}
