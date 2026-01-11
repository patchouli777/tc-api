package category

import (
	"context"
	"errors"
	"fmt"
	"main/internal/db"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RepositoryImpl struct {
	rdb  *redis.Client
	pool *pgxpool.Pool
	// ids of categories sorted by viewers
	sorted sortedStore
	// categories objects
	categories categoryStore
	// map from category link to id
	linkMap linkMap
}

// TODO: properly handle errors (not found, duplicate, etc)
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
		if err == redis.Nil {
			return nil, errNotFound
		}

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
		return nil, fmt.Errorf("error getting list of categories: %w", err)
	}

	categories, err := r.categories.list(ctx, ids)
	if err != nil {
		return nil, err
	}

	// tx := r.rdb.TxPipeline()

	// ids, err := r.sorted.getTx(ctx, tx, start, count)
	// if err != nil {
	// 	return nil, fmt.Errorf("error getting list of categories: %w", err)
	// }

	// categories, err := r.categories.list(ctx, tx, ids)
	// if err != nil {
	// 	return nil, err
	// }

	// _, err = tx.Exec(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("transaction failed: %w", err)
	// }

	return categories, nil
}

func (r *RepositoryImpl) Create(ctx context.Context, cat CategoryCreate) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring connection: %w", err)
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	tx, err := conn.Begin(ctx)
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
		if errors.Is(err, db.ErrDuplicateKey) {
			return errAlreadyExists
		}

		return fmt.Errorf("error creating category: %w", err)
	}

	addedTags, err := qtx.CategoryAddTags(ctx, db.CategoryAddTagsParams{
		Column1: category.ID,
		Column2: cat.Tags,
	})
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	tags := make([]CategoryTag, len(addedTags))
	for i, t := range addedTags {
		tags[i] = CategoryTag{Id: t.TagID, Name: t.TagName}
	}

	err = r.addCategory(ctx, Category{
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

func (r *RepositoryImpl) UpdateViewersById(ctx context.Context, id int, viewers int) error {
	idStr := strconv.Itoa(id)
	err := r.categories.updateViewers(ctx, idStr, viewers)
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

	err = r.categories.updateViewers(ctx, idStr, viewers)
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) UpdateById(ctx context.Context, id int32, cat CategoryUpdate) error {
	return errors.New("not implemented")
}

// TODO: rollback redis transaction =)
// TODO: maybe its better to search by link in postgres?
func (r *RepositoryImpl) UpdateByLink(ctx context.Context, link string, upd CategoryUpdate) error {
	id, err := r.linkMap.get(ctx, link)
	if err != nil {
		return err
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring connection: %w", err)
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	err = q.Update(ctx, db.CategoryUpdateParams{
		ID:           int32(idInt),
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

	if upd.Tags.Explicit {
		int32Ids := make([]int32, len(upd.Tags.Value))

		for i, v := range upd.Tags.Value {
			int32Ids[i] = int32(v)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		qtx := q.queries.WithTx(tx)
		err = qtx.CategoryDeleteTags(ctx, int32(idInt))
		if err != nil {
			return err
		}

		_, err = qtx.CategoryAddTags(ctx, db.CategoryAddTagsParams{
			Column1: int32(idInt),
			Column2: int32Ids,
		})
		if err != nil {
			return err
		}

		tx.Commit(ctx)
	}

	tx := r.rdb.TxPipeline()

	if upd.Link.Explicit {
		r.categories.updateFieldTx(ctx, tx, id, "link", upd.Link.Value)
	}

	if upd.IsSafe.Explicit {
		r.categories.updateFieldTx(ctx, tx, id, "is_safe", upd.IsSafe.Value)
	}

	if upd.Name.Explicit {
		r.categories.updateFieldTx(ctx, tx, id, "name", upd.Name.Value)
	}

	if upd.Thumbnail.Explicit {
		r.categories.updateFieldTx(ctx, tx, id, "thumbnail", upd.Thumbnail.Value)
	}

	if upd.Tags.Explicit {
		r.categories.updateFieldTx(ctx, tx, id, "tags", upd.Tags)
	}

	_, err = tx.Exec(ctx)
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

func (r *RepositoryImpl) DeleteById(ctx context.Context, id int32) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring connection: %w", err)
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	err = q.Delete(ctx, id)
	if err != nil {
		return err
	}

	idString := strconv.FormatInt(int64(id), 10)

	tx := r.rdb.TxPipeline()

	err = r.categories.deleteTx(ctx, tx, idString)
	if err != nil {
		return err
	}

	err = r.sorted.deleteTx(ctx, tx, idString)
	if err != nil {
		return err
	}

	err = r.linkMap.deleteTx(ctx, tx, idString)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx)
	if err != nil {
		return fmt.Errorf("transaction execution failed: %w", err)
	}

	return nil
}

func (r *RepositoryImpl) DeleteByLink(ctx context.Context, link string) error {
	return errors.New("not implemented")
}

// TODO: rollback
func (r *RepositoryImpl) addCategory(ctx context.Context, cat Category) error {
	tx := r.rdb.TxPipeline()

	err := r.categories.addTx(ctx, tx, cat)
	if err != nil {
		return err
	}

	err = r.sorted.addTx(ctx, tx, int(cat.Viewers), strconv.Itoa(int(cat.Id)))
	if err != nil {
		return err
	}

	err = r.linkMap.addTx(ctx, tx, cat.Link, int(cat.Id))
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx)
	if err != nil {
		return fmt.Errorf("transaction execution failed: %w", err)
	}

	return nil
}
