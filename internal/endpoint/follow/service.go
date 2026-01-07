package follow

import (
	"context"
	"fmt"
	"main/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: idempotency
type ServiceImpl struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *ServiceImpl {
	return &ServiceImpl{pool: pool}
}

func (s ServiceImpl) IsFollower(ctx context.Context, follower, followed string) (bool, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	_, err = q.Select(ctx, db.FollowSelectParams{
		Name:   follower,
		Name_2: followed,
	})
	if err != nil {
		return false, fmt.Errorf("failed to determine whether user is follower. %w", err)
	}

	return true, nil
}

func (s ServiceImpl) List(ctx context.Context, follower string) ([]FollowerListItem, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	list, err := q.SelectMany(ctx, follower)
	if err != nil {
		return nil, fmt.Errorf("failed to get following list. %w", err)
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

func (s ServiceImpl) ListExtended(ctx context.Context, follower string) ([]FollowingListExtendedItem, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	list, err := q.SelectManyExtended(ctx, follower)
	if err != nil {
		return nil, fmt.Errorf("failed to get extended following list. %w", err)
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

func (s ServiceImpl) Follow(ctx context.Context, follower, followed string) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	err = q.Insert(ctx, db.FollowInsertParams{
		Column1: pgtype.Text{String: follower, Valid: true},
		Column2: pgtype.Text{String: followed, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to follow. %w", err)
	}

	return nil
}

func (s ServiceImpl) Unfollow(ctx context.Context, unfollower, unfollowed string) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	err = q.Delete(ctx, db.FollowDeleteParams{
		Name:   unfollower,
		Name_2: unfollowed,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow. %w", err)
	}

	return nil
}
