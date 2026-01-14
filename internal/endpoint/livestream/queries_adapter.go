package livestream

import (
	"context"
	"main/internal/db"
)

type QueriesAdapter struct {
	queries *db.Queries
}

func (q *QueriesAdapter) Delete(ctx context.Context, name string) error {
	return q.queries.LivestreamDelete(ctx, name)
}

func (q *QueriesAdapter) Insert(ctx context.Context, arg db.LivestreamInsertParams) (db.LivestreamInsertRow, error) {
	return q.queries.LivestreamInsert(ctx, arg)
}

func (q *QueriesAdapter) Update(ctx context.Context, arg db.LivestreamUpdateParams) (db.LivestreamUpdateRow, error) {
	return q.queries.LivestreamUpdate(ctx, arg)
}

func (q *QueriesAdapter) UpdateViewers(ctx context.Context, arg db.LivestreamUpdateViewersParams) error {
	return q.queries.LivestreamUpdateViewers(ctx, arg)
}
