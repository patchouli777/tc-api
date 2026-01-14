package category

import (
	"context"
	"errors"
	"main/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueriesAdapter struct {
	queries *db.Queries
}

// TODO: sentinel
func (q *QueriesAdapter) Select(ctx context.Context, id int32) (db.TcCategory, error) {
	return q.queries.CategorySelect(ctx, id)
}

func (q *QueriesAdapter) SelectMany(ctx context.Context, arg db.CategorySelectManyParams) ([]db.CategorySelectManyRow, error) {
	return q.queries.CategorySelectMany(ctx, arg)
}

func (q *QueriesAdapter) Update(ctx context.Context, arg db.CategoryUpdateParams) error {
	// "returning" in the underlying query is needed to make sure pgx.ErrNoRows is returned
	_, err := q.queries.CategoryUpdate(ctx, arg)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errNotFound
		}
		return err
	}

	return nil
}

func (q *QueriesAdapter) UpdateByLink(ctx context.Context, arg db.CategoryUpdateByLinkParams) (db.TcCategory, error) {
	updated, err := q.queries.CategoryUpdateByLink(ctx, arg)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.TcCategory{}, errNotFound
		}

		return db.TcCategory{}, err
	}

	return updated, nil
}

// TODO: maybe split query
func (q *QueriesAdapter) UpdateTags(ctx context.Context, categoryId int32, tagsIds []int32) ([]db.CategoryAddTagsRow, error) {
	return q.queries.CategoryAddTags(ctx, db.CategoryAddTagsParams{
		Column1: categoryId,
		Column2: tagsIds,
	})
}

func (q *QueriesAdapter) Insert(ctx context.Context, arg db.CategoryInsertParams) (*db.TcCategory, error) {
	c, err := q.queries.CategoryInsert(ctx, arg)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.CodeUniqueConstraint {
				return nil, errAlreadyExists
			}
		}

		return nil, err
	}

	return &c, nil
}

func (q *QueriesAdapter) Delete(ctx context.Context, id int32) error {
	// "returning" in the underlying query is needed to make sure pgx.ErrNoRows is returned
	_, err := q.queries.CategoryDelete(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errNotFound
		}

		return err
	}

	return nil
}

func (q *QueriesAdapter) DeleteByLink(ctx context.Context, link string) (db.TcCategory, error) {
	deleted, err := q.queries.CategoryDeleteByLink(ctx, link)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.TcCategory{}, errNotFound
		}

		return db.TcCategory{}, err
	}

	return deleted, nil
}

func (q *QueriesAdapter) DeleteTags(ctx context.Context, id int32) error {
	return q.queries.CategoryDeleteTags(ctx, id)
}
