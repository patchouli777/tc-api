package follow

import (
	"context"
	"fmt"
	"main/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{pool: pool}
}

func (r *RepositoryImpl) IsFollower(ctx context.Context, follower, followed string) (bool, error) {
	q := QueriesAdapter{queries: db.New(r.pool)}

	_, err := q.Select(ctx, db.FollowSelectParams{
		Name:   follower,
		Name_2: followed,
	})
	if err != nil {
		return false, fmt.Errorf("failed to determine whether user is follower: %w", err)
	}

	return true, nil
}

func (r *RepositoryImpl) List(ctx context.Context, follower string) ([]FollowerListItem, error) {
	q := QueriesAdapter{queries: db.New(r.pool)}

	list, err := q.SelectMany(ctx, follower)
	if err != nil {
		return nil, fmt.Errorf("failed to get following list: %w", err)
	}

	following := make([]FollowerListItem, len(list))
	for i, f := range list {
		following[i] = FollowerListItem{
			Name:   f.Name,
			Avatar: f.Avatar.String,
		}
	}

	return following, nil
}

func (r *RepositoryImpl) ListExtended(ctx context.Context, follower string) ([]FollowingListExtendedItem, error) {
	q := QueriesAdapter{queries: db.New(r.pool)}

	list, err := q.SelectManyExtended(ctx, follower)
	if err != nil {
		return nil, fmt.Errorf("failed to get extended following list: %w", err)
	}

	following := make([]FollowingListExtendedItem, len(list))
	for i, f := range list {
		following[i] = FollowingListExtendedItem{
			Name:     f.Following.String,
			Avatar:   f.Avatar.String,
			Viewers:  f.Viewers.Int32,
			Title:    f.Title.String,
			Category: f.Category.String,
			IsLive:   f.IsLive.Bool,
		}
	}

	return following, nil
}

func (r *RepositoryImpl) Follow(ctx context.Context, follower, followed string) error {
	q := QueriesAdapter{queries: db.New(r.pool)}

	err := q.Insert(ctx, db.FollowInsertParams{
		Column1: pgtype.Text{String: follower, Valid: true},
		Column2: pgtype.Text{String: followed, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to follow: %w", err)
	}

	return nil
}

func (r *RepositoryImpl) Unfollow(ctx context.Context, unfollower, unfollowed string) error {
	q := QueriesAdapter{queries: db.New(r.pool)}

	err := q.Delete(ctx, db.FollowDeleteParams{
		Name:   unfollower,
		Name_2: unfollowed,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	return nil
}
