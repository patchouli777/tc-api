package storage

import (
	"context"
	"fmt"
	"strconv"

	d "main/internal/category/domain"

	"github.com/redis/go-redis/v9"
)

type cache struct {
	rdb *redis.Client
	// ids of categories sorted by viewers
	sorted *sortedStore
	// categories objects
	categories *categoryStore
	// map from category link to id
	linkMap *linkMap
}

// TODO: properly handle errors (not found, duplicate, etc) + check tx for consistency
func newCache(rdb *redis.Client) *cache {
	sorted := sortedStore{rdb: rdb}
	categories := categoryStore{rdb: rdb}
	linkMap := linkMap{rdb: rdb}

	return &cache{rdb: rdb,
		sorted:     &sorted,
		categories: &categories,
		linkMap:    &linkMap,
	}
}

func (r *cache) get(ctx context.Context, id int) (*d.Category, error) {
	idStr := strconv.Itoa(id)
	category, err := r.categories.get(ctx, idStr)
	if err != nil {
		if err == redis.Nil {
			return nil, d.ErrNotFound
		}

		return nil, err
	}

	return category, nil
}

func (r *cache) getByLink(ctx context.Context, link string) (*d.Category, error) {
	id, err := r.linkMap.get(ctx, link)
	if err != nil {
		if err == redis.Nil {
			return nil, d.ErrNotFound
		}

		return nil, err
	}

	category, err := r.categories.get(ctx, id)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (r *cache) add(ctx context.Context, cat d.Category) error {
	cmds, err := r.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		r.linkMap.addTx(ctx, p, cat.Link, int(cat.Id))
		r.sorted.addTx(ctx, p, int(cat.Viewers), strconv.Itoa(int(cat.Id)))
		r.categories.addTx(ctx, p, cat)
		return nil
	})

	if err != nil {
		return fmt.Errorf("pipeline failed: %w", err)
	}

	for i, cmd := range cmds {
		if cmd.Err() != nil {
			return fmt.Errorf("cmd %d failed: %w", i, cmd.Err())
		}
	}

	return nil

}

func (r *cache) list(ctx context.Context, f d.CategoryFilter) ([]d.Category, error) {
	start := (int64(f.Page) - 1) * int64(f.Count)
	count := int64(f.Count)

	ids, err := r.sorted.get(ctx, start, count)
	if err != nil {
		return nil, err
	}

	categories, err := r.categories.list(ctx, ids)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *cache) update(ctx context.Context, id int, upd d.CategoryUpdate) error {
	idStr := strconv.Itoa(id)
	return r.categories.update(ctx, idStr, map[string]any{
		"link":      upd.Link.Value,
		"is_safe":   upd.IsSafe.Value,
		"name":      upd.Name.Value,
		"thumbnail": upd.Thumbnail.Value,
		"tags":      upd.Tags,
	})
}

func (r *cache) updateViewers(ctx context.Context, id int, viewers int) error {
	idStr := strconv.Itoa(id)
	err := r.categories.updateViewers(ctx, idStr, viewers)
	if err != nil {
		return err
	}

	err = r.sorted.updateViewers(ctx, idStr, int32(viewers))
	if err != nil {
		return err
	}

	return nil
}

func (r *cache) delete(ctx context.Context, id int) error {
	idStr := strconv.Itoa(id)

	cmds, err := r.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		r.categories.deleteTx(ctx, p, idStr)
		r.sorted.deleteTx(ctx, p, idStr)
		r.linkMap.deleteTx(ctx, p, idStr)
		return nil
	})

	if err != nil {
		return fmt.Errorf("pipeline failed: %w", err)
	}

	for i, cmd := range cmds {
		if cmd.Err() != nil {
			return fmt.Errorf("cmd %d failed: %w", i, cmd.Err())
		}
	}

	return nil
}
