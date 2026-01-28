package storage

import (
	"context"
	"twitchy-api/internal/external/db"
)

type queriesAdapter struct {
	queries *db.Queries
}

func (q *queriesAdapter) Delete(ctx context.Context, id int) error {
	return q.queries.LivestreamDelete(ctx, int32(id))
}

func (q *queriesAdapter) Insert(ctx context.Context, username string) (db.LivestreamInsertRow, error) {
	return q.queries.LivestreamInsert(ctx, username)
}

func (q *queriesAdapter) Update(ctx context.Context, arg db.LivestreamUpdateParams) (db.LivestreamUpdateRow, error) {
	return q.queries.LivestreamUpdate(ctx, arg)
}

func (q *queriesAdapter) UpdateViewers(ctx context.Context, arg db.LivestreamUpdateViewersParams) error {
	return q.queries.LivestreamUpdateViewers(ctx, arg)
}
