package category

import (
	"context"
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

func (r *categoryStore) get(ctx context.Context, id string) (*Category, error) {
	var category Category
	err := r.rdb.HGetAll(ctx, r.key(id)).Scan(&category)
	return &category, err
}

func (r *categoryStore) list(ctx context.Context, ids []string) ([]Category, error) {
	pipe := r.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))

	for i, id := range ids {
		cmds[i] = pipe.HGetAll(ctx, r.key(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	categories := make([]Category, len(ids))
	for i, cmd := range cmds {
		var category Category

		err := cmd.Scan(&category)
		if err != nil {
			return nil, err
		}

		categories[i] = category
	}

	return categories, nil
}

func (r *categoryStore) key(id string) string {
	return fmt.Sprintf("categories:%s", id)
}

func (r *categoryStore) update(ctx context.Context, id string, viewers int) error {
	return r.rdb.HSet(ctx, r.key(id), "viewers", viewers).Err()
}

func (r *categoryStore) delete(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, r.key(id)).Err()
}
