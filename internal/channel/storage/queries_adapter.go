package storage

import (
	"context"
	"twitchy-api/internal/external/db"
)

type queriesAdapter struct {
	queries *db.Queries
}

func (q *queriesAdapter) Select(ctx context.Context, name string) (db.ChannelSelectRow, error) {
	return q.queries.ChannelSelect(ctx, name)
}
