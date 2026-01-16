package follow

import (
	"context"
	"errors"
	"fmt"
	"main/internal/external/db"

	"github.com/jackc/pgx/v5/pgconn"
)

type QueriesAdapter struct {
	queries *db.Queries
}

func (q *QueriesAdapter) Select(ctx context.Context, arg db.FollowSelectParams) (db.FollowSelectRow, error) {
	return q.queries.FollowSelect(ctx, arg)
}

func (q *QueriesAdapter) SelectMany(ctx context.Context, name string) ([]db.FollowSelectManyRow, error) {
	return q.queries.FollowSelectMany(ctx, name)
}

func (q *QueriesAdapter) SelectManyExtended(ctx context.Context, name string) ([]db.FollowSelectManyExtendedRow, error) {
	return q.queries.FollowSelectManyExtended(ctx, name)
}

func (q *QueriesAdapter) SelectUserId(ctx context.Context, name string) (int32, error) {
	return q.queries.FollowSelectUserId(ctx, name)
}

func (q *QueriesAdapter) Insert(ctx context.Context, arg db.FollowInsertParams) error {
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

func (q *QueriesAdapter) Delete(ctx context.Context, arg db.FollowDeleteParams) error {
	err := q.queries.FollowDelete(ctx, arg)

	fmt.Println(err)

	return err
}
