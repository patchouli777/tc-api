package user

import (
	"context"
	"errors"
	"main/internal/external/db"
	d "main/internal/user/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queriesAdapter struct {
	queries *db.Queries
}

func (q *queriesAdapter) Select(ctx context.Context, id int32) (db.UserSelectRow, error) {
	res, err := q.queries.UserSelect(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.UserSelectRow{}, d.ErrNotFound
		}

		return db.UserSelectRow{}, err
	}

	return res, err
}

func (q *queriesAdapter) SelectByUsername(ctx context.Context, username string) (db.UserSelectByUsernameRow, error) {
	res, err := q.queries.UserSelectByUsername(ctx, username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.UserSelectByUsernameRow{}, d.ErrNotFound
		}

		return db.UserSelectByUsernameRow{}, err
	}

	return res, err
}

func (q *queriesAdapter) Insert(ctx context.Context, p db.UserInsertParams) error {
	_, err := q.queries.UserInsert(ctx, p)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.CodeUniqueConstraint {
				return d.ErrAlreadyExists
			}
		}

		return err
	}

	return err
}

func (q *queriesAdapter) Update(ctx context.Context, arg db.UserUpdateParams) error {
	err := q.queries.UserUpdate(ctx, arg)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.CodeUniqueConstraint {
				return d.ErrAlreadyExists
			}
		}

		return err
	}

	return err
}

func (q *queriesAdapter) Delete(ctx context.Context, id int32) error {
	return q.queries.UserDelete(ctx, id)
}
