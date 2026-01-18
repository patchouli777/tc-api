package storage

import (
	"context"
	"errors"
	"fmt"
	d "main/internal/category/domain"
	"main/internal/external/db"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RepositoryImpl struct {
	pool  *pgxpool.Pool
	cache *cache
}

func NewRepo(rdb *redis.Client, pool *pgxpool.Pool) *RepositoryImpl {
	cache := newCache(rdb)
	return &RepositoryImpl{pool: pool, cache: cache}
}

func (r *RepositoryImpl) Get(ctx context.Context, id int) (*d.Category, error) {
	return r.cache.get(ctx, id)
}

func (r *RepositoryImpl) GetByLink(ctx context.Context, link string) (*d.Category, error) {
	return r.cache.getByLink(ctx, link)
}

func (r *RepositoryImpl) List(ctx context.Context, f d.CategoryFilter) ([]d.Category, error) {
	return r.cache.list(ctx, f)
}

func (r *RepositoryImpl) Create(ctx context.Context, cat d.CategoryCreate) error {
	q := queriesAdapter{queries: db.New(r.pool)}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := q.queries.WithTx(tx)

	category, err := qtx.CategoryInsert(ctx, db.CategoryInsertParams{
		Name:  cat.Name,
		Link:  cat.Link,
		Image: cat.Thumbnail,
	})
	if err != nil {
		// TODO: this should be in adapter probably
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.CodeUniqueConstraint {
				return d.ErrAlreadyExists
			}
		}

		return fmt.Errorf("error creating category: %w", err)
	}

	tagsInt32 := make([]int32, len(cat.Tags))
	for i, t := range cat.Tags {
		tagsInt32[i] = int32(t)
	}

	addedTags, err := qtx.CategoryAddTags(ctx, db.CategoryAddTagsParams{
		Column1: category.ID,
		Column2: tagsInt32,
	})
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	tags := make([]d.CategoryTag, len(addedTags))
	for i, t := range addedTags {
		tags[i] = d.CategoryTag{Id: t.TagID, Name: t.TagName}
	}

	err = r.cache.add(ctx, d.Category{
		Id:        category.ID,
		IsSafe:    category.IsSafe,
		Thumbnail: category.Image,
		Name:      category.Name,
		Link:      category.Link,
		Viewers:   0,
		Tags:      tags,
	})
	if err != nil {
		return err
	}

	return nil
}

// TODO: currently update uses 2 db connections to be performed:
// one for q.Update() and another one for transaction in updateTags()
// should be 1
func (r *RepositoryImpl) Update(ctx context.Context, id int32, upd d.CategoryUpdate) error {
	q := queriesAdapter{queries: db.New(r.pool)}

	err := q.Update(ctx, db.CategoryUpdateParams{
		ID:           id,
		NameDoUpdate: upd.Name.Explicit,
		Name:         upd.Name.Value,

		LinkDoUpdate: upd.Link.Explicit,
		Link:         upd.Link.Value,

		ImageDoUpdate: upd.Thumbnail.Explicit,
		Image:         upd.Thumbnail.Value,

		IsSafeDoUpdate: upd.IsSafe.Explicit,
		IsSafe:         upd.IsSafe.Value,
	})
	if err != nil {
		return err
	}

	tagsInt32 := make([]int32, len(upd.Tags.Value))
	for i, t := range upd.Tags.Value {
		tagsInt32[i] = int32(t)
	}

	if upd.Tags.Explicit {
		err = r.updateTags(ctx, &q, id, tagsInt32)
		if err != nil {
			return err
		}
	}

	err = r.cache.update(ctx, int(id), upd)
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) UpdateByLink(ctx context.Context, link string, upd d.CategoryUpdate) error {
	q := queriesAdapter{queries: db.New(r.pool)}

	updated, err := q.UpdateByLink(ctx, db.CategoryUpdateByLinkParams{
		Link: link,

		NameDoUpdate: upd.Name.Explicit,
		Name:         upd.Name.Value,

		LinkDoUpdate: upd.Link.Explicit,
		LinkUpd:      upd.Link.Value,

		ImageDoUpdate: upd.Thumbnail.Explicit,
		Image:         upd.Thumbnail.Value,

		IsSafeDoUpdate: upd.IsSafe.Explicit,
		IsSafe:         upd.IsSafe.Value,
	})
	if err != nil {
		return fmt.Errorf("category update failed:%w", err)
	}

	tagsInt32 := make([]int32, len(upd.Tags.Value))
	for i, t := range upd.Tags.Value {
		tagsInt32[i] = int32(t)
	}

	if upd.Tags.Explicit {
		err = r.updateTags(ctx, &q, updated.ID, tagsInt32)
		if err != nil {
			return fmt.Errorf("tags update failed:%w", err)
		}
	}

	err = r.cache.update(ctx, int(updated.ID), upd)
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) UpdateViewers(ctx context.Context, id int, viewers int) error {
	return r.cache.updateViewers(ctx, id, viewers)
}

func (r *RepositoryImpl) updateTags(ctx context.Context, q *queriesAdapter, id int32, tags []int32) error {
	int32Ids := make([]int32, len(tags))

	for i, v := range tags {
		int32Ids[i] = int32(v)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := q.queries.WithTx(tx)
	err = qtx.CategoryDeleteTags(ctx, id)
	if err != nil {
		return err
	}

	_, err = qtx.CategoryAddTags(ctx, db.CategoryAddTagsParams{
		Column1: id,
		Column2: int32Ids,
	})
	if err != nil {
		return err
	}

	tx.Commit(ctx)

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, id int32) error {
	q := queriesAdapter{queries: db.New(r.pool)}

	err := q.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = r.cache.delete(ctx, int(id))
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) DeleteByLink(ctx context.Context, link string) error {
	q := queriesAdapter{queries: db.New(r.pool)}

	deleted, err := q.DeleteByLink(ctx, link)
	if err != nil {
		return err
	}

	err = r.cache.delete(ctx, int(deleted.ID))
	if err != nil {
		return err
	}

	return nil
}
