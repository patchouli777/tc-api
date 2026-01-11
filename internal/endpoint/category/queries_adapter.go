package category

import (
	"context"
	"errors"
	"main/internal/db"

	"github.com/jackc/pgx/v5/pgconn"
)

type QueriesAdapter struct {
	queries *db.Queries
}

func (q *QueriesAdapter) SelectMany(ctx context.Context, arg db.CategorySelectManyParams) ([]db.CategorySelectManyRow, error) {
	return q.queries.CategorySelectMany(ctx, arg)
}

func (q *QueriesAdapter) Update(ctx context.Context, arg db.CategoryUpdateParams) error {
	return q.queries.CategoryUpdate(ctx, arg)
}

func (q *QueriesAdapter) Insert(ctx context.Context, arg db.CategoryInsertParams) (*db.TcCategory, error) {
	c, err := q.queries.CategoryInsert(ctx, arg)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, db.ErrDuplicateKey
			}
		}

		return nil, err
	}

	return &c, nil
}

func (q *QueriesAdapter) Select(ctx context.Context, id int32) (db.TcCategory, error) {
	return q.queries.CategorySelect(ctx, id)
}

func (q *QueriesAdapter) Delete(ctx context.Context, id int32) error {
	return q.queries.CategoryDelete(ctx, id)
}

func (q *QueriesAdapter) DeleteTags(ctx context.Context, id int32) error {
	return q.queries.CategoryDeleteTags(ctx, id)
}

// TODO: maybe split query
func (q *QueriesAdapter) UpdateTags(ctx context.Context, categoryId int32, tagsIds []int32) ([]db.CategoryAddTagsRow, error) {
	return q.queries.CategoryAddTags(ctx, db.CategoryAddTagsParams{
		Column1: categoryId,
		Column2: tagsIds,
	})
}
