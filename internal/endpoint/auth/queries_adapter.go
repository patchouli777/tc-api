package auth

import (
	"context"
	"errors"
	"main/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueriesAdapter struct {
	Pool *pgxpool.Pool
}

func NewAdapter(pool *pgxpool.Pool) *QueriesAdapter {
	return &QueriesAdapter{Pool: pool}
}

func (q *QueriesAdapter) Select(ctx context.Context, arg db.AuthSelectUserParams) (*db.AuthSelectUserRow, error) {
	queries := db.New(q.Pool)
	res, err := queries.AuthSelectUser(ctx, arg)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errNotFound
	}

	return &res, nil
}

func (q *QueriesAdapter) Insert(ctx context.Context, arg db.AuthInsertUserParams) error {
	queries := db.New(q.Pool)
	err := queries.AuthInsertUser(ctx, arg)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == db.CodeUniqueConstraint {
			return errAlreadyExists
		}
	}

	return err
}
