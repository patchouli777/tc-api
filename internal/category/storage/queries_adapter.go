package storage

import (
	"context"
	"errors"
	d "twitchy-api/internal/category/domain"
	"twitchy-api/internal/external/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queriesAdapter struct {
	queries *db.Queries
}

// TODO: sentinel
func (q *queriesAdapter) Select(ctx context.Context, id int32) (db.TcCategory, error) {
	return q.queries.CategorySelect(ctx, id)
}

func (q *queriesAdapter) SelectMany(ctx context.Context, arg db.CategorySelectManyParams) ([]db.CategorySelectManyRow, error) {
	return q.queries.CategorySelectMany(ctx, arg)
}

func (q *queriesAdapter) Update(ctx context.Context, arg db.CategoryUpdateParams) error {
	// "returning" in the underlying query is needed to make sure pgx.ErrNoRows is returned
	_, err := q.queries.CategoryUpdate(ctx, arg)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.ErrNotFound
		}
		return err
	}

	return nil
}

func (q *queriesAdapter) UpdateByLink(ctx context.Context, arg db.CategoryUpdateByLinkParams) (db.TcCategory, error) {
	updated, err := q.queries.CategoryUpdateByLink(ctx, arg)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.TcCategory{}, d.ErrNotFound
		}

		return db.TcCategory{}, err
	}

	return updated, nil
}

// TODO: maybe split query
func (q *queriesAdapter) UpdateTags(ctx context.Context, categoryId int32, tagsIds []int32) ([]db.CategoryAddTagsRow, error) {
	return q.queries.CategoryAddTags(ctx, db.CategoryAddTagsParams{
		Column1: categoryId,
		Column2: tagsIds,
	})
}

func (q *queriesAdapter) Insert(ctx context.Context, arg db.CategoryInsertParams) (*db.TcCategory, error) {
	c, err := q.queries.CategoryInsert(ctx, arg)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.CodeUniqueConstraint {
				return nil, d.ErrAlreadyExists
			}
		}

		return nil, err
	}

	return &c, nil
}

func (q *queriesAdapter) Delete(ctx context.Context, id int32) error {
	// "returning" in the underlying query is needed to make sure pgx.ErrNoRows is returned
	_, err := q.queries.CategoryDelete(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.ErrNotFound
		}

		return err
	}

	return nil
}

// TODO: return nil if no rows?
func (q *queriesAdapter) DeleteByLink(ctx context.Context, link string) (db.TcCategory, error) {
	deleted, err := q.queries.CategoryDeleteByLink(ctx, link)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.TcCategory{}, d.ErrNotFound
		}

		return db.TcCategory{}, err
	}

	return deleted, nil
}

func (q *queriesAdapter) DeleteTags(ctx context.Context, id int32) error {
	return q.queries.CategoryDeleteTags(ctx, id)
}
