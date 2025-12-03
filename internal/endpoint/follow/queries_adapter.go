package follow

import (
	"context"
	"main/internal/db"
)

type QueriesAdapter struct {
	queries *db.Queries
}

func (q *QueriesAdapter) Select(ctx context.Context, arg db.FollowSelectParams) (db.FollowSelectRow, error) {
	return q.queries.FollowSelect(ctx, arg)
}

func (q *QueriesAdapter) Delete(ctx context.Context, arg db.FollowDeleteParams) error {
	return q.queries.FollowDelete(ctx, arg)
}

func (q *QueriesAdapter) SelectUserId(ctx context.Context, name string) (int32, error) {
	return q.queries.FollowSelectUserId(ctx, name)
}

func (q *QueriesAdapter) Insert(ctx context.Context, arg db.FollowInsertParams) error {
	return q.queries.FollowInsert(ctx, arg)
}

func (q *QueriesAdapter) SelectMany(ctx context.Context, name string) ([]db.FollowSelectManyRow, error) {
	return q.queries.FollowSelectMany(ctx, name)
}

func (q *QueriesAdapter) SelectManyExtended(ctx context.Context, name string) ([]db.FollowSelectManyExtendedRow, error) {
	return q.queries.FollowSelectManyExtended(ctx, name)
}
