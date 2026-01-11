package category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// store categories retrievable by id
type categoryStore struct {
	rdb *redis.Client
}

func (r *categoryStore) add(ctx context.Context, cat Category) error {
	idStr := strconv.Itoa(int(cat.Id))
	return r.rdb.HSet(ctx, r.key(idStr), cat).Err()
}

func (r *categoryStore) addTx(ctx context.Context, tx redis.Pipeliner, cat Category) error {
	idStr := strconv.Itoa(int(cat.Id))
	return tx.HSet(ctx, r.key(idStr), cat).Err()
}

// cmd must be a result of HGetAll
func (r *categoryStore) parseCategory(ctx context.Context, cmd *redis.MapStringStringCmd) (*Category, error) {
	res, err := cmd.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errNotFound
		}

		return nil, err
	}

	if len(res) == 0 {
		return nil, errNotFound
	}

	var category Category
	err = cmd.Scan(&category)
	if err != nil {
		return nil, err
	}

	if tagsJSON, ok := res["tags"]; ok {
		json.Unmarshal([]byte(tagsJSON), &category.Tags)
	}

	return &category, nil
}

// TODO: is there any better way???
// redis marshaling doesn't work with slices and it is big bad
func (r *categoryStore) get(ctx context.Context, id string) (*Category, error) {
	cmd := r.rdb.HGetAll(ctx, r.key(id))
	return r.parseCategory(ctx, cmd)
}

func (r *categoryStore) getTx(ctx context.Context, tx redis.Pipeliner, id string) (*Category, error) {
	var category Category

	err := tx.HGetAll(ctx, r.key(id)).Scan(&category)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errNotFound
		}

		return nil, err
	}

	return &category, nil
}

func (r *categoryStore) list(ctx context.Context, ids []string) ([]Category, error) {
	pipe := r.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))

	for i, id := range ids {
		cmds[i] = pipe.HGetAll(ctx, r.key(id))
	}

	_, err := pipe.Exec(ctx)
	// TODO: why != redis.Nil?
	if err != nil && err != redis.Nil {
		return nil, err
	}

	categories := make([]Category, len(ids))
	for i, cmd := range cmds {
		category, err := r.parseCategory(ctx, cmd)
		if err != nil {
			return nil, err
		}

		categories[i] = *category
	}

	return categories, nil
}

func (r *categoryStore) key(id string) string {
	return fmt.Sprintf("categories:%s", id)
}

func (r *categoryStore) updateViewers(ctx context.Context, id string, viewers int) error {
	return r.rdb.HSet(ctx, r.key(id), "viewers", viewers).Err()
}

func (r *categoryStore) updateField(ctx context.Context, id, field string, val any) error {
	return r.rdb.HSet(ctx, r.key(id), field, val).Err()
}

func (r *categoryStore) updateFieldTx(ctx context.Context, tx redis.Pipeliner, id, field string, val any) *redis.IntCmd {
	return tx.HSet(ctx, r.key(id), field, val)
}

func (r *categoryStore) delete(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, r.key(id)).Err()
}

func (r *categoryStore) deleteTx(ctx context.Context, tx redis.Pipeliner, id string) error {
	return tx.Del(ctx, r.key(id)).Err()
}
