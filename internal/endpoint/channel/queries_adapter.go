package channel

import (
	"context"
	"main/internal/db"
)

type QueriesAdapter struct {
	queries *db.Queries
}

func (q *QueriesAdapter) Select(ctx context.Context, name string) (db.ChannelSelectRow, error) {
	return q.queries.ChannelSelect(ctx, name)
}
