package category

import (
	"context"
	"errors"
	"fmt"
	"main/internal/db"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RepositoryImpl struct {
	rdb        *redis.Client
	pool       *pgxpool.Pool
	sorted     sortedStore
	categories categoryStore
	linkMap    linkMap
}

func NewRepo(rdb *redis.Client, pool *pgxpool.Pool) *RepositoryImpl {
	sorted := sortedStore{rdb: rdb}
	categories := categoryStore{rdb: rdb}
	linkMap := linkMap{rdb: rdb}

	return &RepositoryImpl{rdb: rdb,
		pool:       pool,
		sorted:     sorted,
		categories: categories,
		linkMap:    linkMap,
	}
}

func (r *RepositoryImpl) GetById(ctx context.Context, id int) (*Category, error) {
	idStr := strconv.Itoa(id)
	category, err := r.categories.get(ctx, idStr)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// TODO: try pipeline
func (r *RepositoryImpl) GetByLink(ctx context.Context, link string) (*Category, error) {
	id, err := r.linkMap.get(ctx, link)
	if err != nil {
		return nil, err
	}

	category, err := r.categories.get(ctx, id)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// TODO: try pipeline
func (r *RepositoryImpl) List(ctx context.Context, f CategoryFilter) ([]Category, error) {
	start := (int64(f.Page) - 1) * int64(f.Count)
	count := int64(f.Count)

	ids, err := r.sorted.get(ctx, start, count)
	if err != nil {
		return nil, fmt.Errorf("error getting list of categories: %v", err)
	}

	categories, err := r.categories.list(ctx, ids)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// TODO: добавить теги
func (r *RepositoryImpl) Create(ctx context.Context, cat CategoryCreate) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring connection: %v", err)
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	inserted, err := q.Insert(ctx, db.CategoryInsertParams{
		Name:  cat.Name,
		Link:  cat.Link,
		Image: pgtype.Text{String: cat.Thumbnail, Valid: true},
	})
	if err != nil {
		if errors.Is(err, db.ErrDuplicateKey) {
			return ErrCategoryAlreadyExists
		}

		return fmt.Errorf("error creating category: %v", err)
	}

	category := Category{
		Id:        inserted.ID,
		IsSafe:    inserted.IsSafe.Bool,
		Thumbnail: inserted.Image.String,
		Name:      inserted.Name,
		Link:      inserted.Link,
		Viewers:   inserted.Viewers,
		Tags:      []string{},
	}

	err = r.addCategory(ctx, category)
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) UpdateViewersById(ctx context.Context, id int, viewers int) error {
	idStr := strconv.Itoa(id)
	err := r.categories.update(ctx, idStr, viewers)
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) UpdateViewersByLink(ctx context.Context, link string, viewers int) error {
	idStr, err := r.linkMap.get(ctx, link)
	if err != nil {
		return err
	}

	err = r.categories.update(ctx, idStr, viewers)
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) UpdateById(ctx context.Context, id int, cat CategoryUpdate) error {
	return errors.New("not implemented")
}

// TODO: better update (pipeline)
func (r *RepositoryImpl) UpdateByLink(ctx context.Context, link string, upd CategoryUpdate) error {
	cur, err := r.GetByLink(ctx, link)
	if err != nil {
		return err
	}

	if upd.IsSafe != nil {
		cur.IsSafe = *upd.IsSafe
	}

	if upd.Thumbnail != nil {
		cur.Thumbnail = *upd.Thumbnail
	}

	if upd.Name != nil {
		cur.Name = *upd.Name
	}

	if upd.Link != nil {
		cur.Link = *upd.Link
	}

	if upd.Viewers != nil {
		cur.Viewers = *upd.Viewers
	}

	if upd.Tags != nil {
		cur.Tags = *upd.Tags
	}

	if err = r.categories.add(ctx, *cur); err != nil {
		return err
	}

	if err = r.sorted.add(ctx, int(cur.Viewers), strconv.Itoa(int(cur.Id))); err != nil {
		return err
	}

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring connection: %v", err)
	}
	defer conn.Release()

	// q := QueriesAdapter{Q: db.New(conn)}

	return nil
}

func (r *RepositoryImpl) DeleteById(ctx context.Context, id int) error {
	return errors.New("not implemented")
}

func (r *RepositoryImpl) DeleteByLink(ctx context.Context, link string) error {
	return errors.New("not implemented")
}

func (r *RepositoryImpl) addCategory(ctx context.Context, cat Category) error {
	err := r.categories.add(ctx, cat)
	if err != nil {
		return err
	}

	err = r.sorted.add(ctx, int(cat.Viewers), strconv.Itoa(int(cat.Id)))
	if err != nil {
		return err
	}

	err = r.linkMap.add(ctx, cat.Link, int(cat.Id))
	if err != nil {
		return err
	}

	return nil
}
