package storage

import (
	"context"
	"errors"
	d "main/internal/auth/domain"
	"main/internal/external/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queriesAdapter struct {
	queries *db.Queries
}

func (q *queriesAdapter) Select(ctx context.Context, arg db.AuthSelectUserParams) (*db.AuthSelectUserRow, error) {
	res, err := q.queries.AuthSelectUser(ctx, arg)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, d.ErrNotFound
	}

	return &res, nil
}

func (q *queriesAdapter) Insert(ctx context.Context, arg db.AuthInsertUserParams) error {
	err := q.queries.AuthInsertUser(ctx, arg)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == db.CodeUniqueConstraint {
			return d.ErrAlreadyExists
		}
	}

	return err
}
