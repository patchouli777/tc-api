package user

import (
	"context"
	"errors"
	"main/internal/db"

	"github.com/jackc/pgx/v5/pgconn"
)

type QueriesAdapter struct {
	queries *db.Queries
}

func NewDBAdapter(queries *db.Queries) *QueriesAdapter {
	return &QueriesAdapter{queries: queries}
}

func (q *QueriesAdapter) SelectById(ctx context.Context, id int32) (db.UserSelectByIdRow, error) {
	return q.queries.UserSelectById(ctx, id)
}

func (q *QueriesAdapter) SelectByUsername(ctx context.Context, username string) (db.UserSelectByUsernameRow, error) {
	return q.queries.UserSelectByUsername(ctx, username)
}

func (q *QueriesAdapter) Insert(ctx context.Context, p db.UserInsertParams) (int32, error) {
	return q.queries.UserInsert(ctx, p)
}

func (q *QueriesAdapter) Delete(ctx context.Context, id int32) error {
	return q.queries.UserDeleteById(ctx, id)
}

func (q *QueriesAdapter) UpdateById(ctx context.Context, arg db.UserUpdateByIdParams) error {
	err := q.queries.UserUpdateById(ctx, arg)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return db.ErrDuplicateKey
		}
	}

	return err
}
