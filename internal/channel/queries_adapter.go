package channel

import (
	"context"
	"main/internal/external/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QueriesAdapter struct {
	pool *pgxpool.Pool
}

func NewAdapter(pool *pgxpool.Pool) *QueriesAdapter {
	return &QueriesAdapter{pool: pool}
}

func (q *QueriesAdapter) Select(ctx context.Context, name string) (db.ChannelSelectRow, error) {
	conn, err := q.pool.Acquire(ctx)
	if err != nil {
		return db.ChannelSelectRow{}, err
	}
	defer conn.Release()

	queries := db.New(conn)

	return queries.ChannelSelect(ctx, name)
}
