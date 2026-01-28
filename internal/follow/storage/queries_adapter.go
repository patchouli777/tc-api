package follow

import (
	"context"
	"errors"
	"twitchy-api/internal/external/db"

	"github.com/jackc/pgx/v5/pgconn"
)

type queriesAdapter struct {
	queries *db.Queries
}

func (q *queriesAdapter) Select(ctx context.Context, arg db.FollowSelectParams) (db.FollowSelectRow, error) {
	return q.queries.FollowSelect(ctx, arg)
}

func (q *queriesAdapter) SelectMany(ctx context.Context, name string) ([]db.FollowSelectManyRow, error) {
	return q.queries.FollowSelectMany(ctx, name)
}

func (q *queriesAdapter) SelectManyExtended(ctx context.Context, name string) ([]db.FollowSelectManyExtendedRow, error) {
	return q.queries.FollowSelectManyExtended(ctx, name)
}

func (q *queriesAdapter) SelectUserId(ctx context.Context, name string) (int32, error) {
	return q.queries.FollowSelectUserId(ctx, name)
}

func (q *queriesAdapter) Insert(ctx context.Context, arg db.FollowInsertParams) error {
	err := q.queries.FollowInsert(ctx, arg)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.CodeUniqueConstraint {
				return nil
			}
		}

		return err
	}

	return err
}

func (q *queriesAdapter) Delete(ctx context.Context, arg db.FollowDeleteParams) error {
	return q.queries.FollowDelete(ctx, arg)
}
