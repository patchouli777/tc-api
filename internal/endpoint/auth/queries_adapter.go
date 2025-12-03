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
	conn, err := q.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	queries := *db.New(conn)
	res, err := queries.AuthSelectUser(ctx, arg)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, db.ErrNotFound
	}

	return &res, nil
}

func (q *QueriesAdapter) Insert(ctx context.Context, arg db.AuthInsertUserParams) error {
	conn, err := q.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	queries := *db.New(conn)

	err = queries.AuthInsertUser(ctx, arg)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return db.ErrDuplicateKey
		}
	}

	return err
}
