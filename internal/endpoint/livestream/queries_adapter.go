package livestream

import (
	"context"
	"main/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type QueriesAdapter struct {
	queries *db.Queries
}

func (q *QueriesAdapter) Delete(ctx context.Context, name string) error {
	return q.queries.LivestreamDelete(ctx, name)
}

func (q *QueriesAdapter) SelectById(ctx context.Context, name string) (db.LivestreamSelectByIdRow, error) {
	return q.queries.LivestreamSelectById(ctx, name)
}

func (q *QueriesAdapter) List(ctx context.Context, arg db.LivestreamSelectManyParams) ([]db.LivestreamSelectManyRow, error) {
	return q.queries.LivestreamSelectMany(ctx, arg)
}

func (q *QueriesAdapter) ListFromCategory(ctx context.Context, arg db.LivestreamSelectManyFromCategoryParams) ([]db.LivestreamSelectManyFromCategoryRow, error) {
	return q.queries.LivestreamSelectManyFromCategory(ctx, arg)
}

func (q *QueriesAdapter) SelectCategoryLink(ctx context.Context, arg int32) (string, error) {
	return q.queries.LivestreamSelectCategoryLinkById(ctx, arg)
}

func (q *QueriesAdapter) SelectUser(ctx context.Context, name string) (db.LivestreamSelectUserRow, error) {
	return q.queries.LivestreamSelectUser(ctx, name)
}

func (q *QueriesAdapter) SelectUserDetails(ctx context.Context, name string) (pgtype.Text, error) {
	return q.queries.LivestreamSelectUserDetails(ctx, name)
}

func (q *QueriesAdapter) SelectUsersDetails(ctx context.Context, usernames []string) ([]db.LivestreamSelectUsersDetailsRow, error) {
	return q.queries.LivestreamSelectUsersDetails(ctx, usernames)
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
