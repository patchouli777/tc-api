package follow

import (
	"context"
	"log/slog"
	"main/internal/db"
	"main/internal/lib/sl"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: idempotency
type ServiceImpl struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

func NewService(log *slog.Logger, pool *pgxpool.Pool) *ServiceImpl {
	return &ServiceImpl{log: log, pool: pool}
}

func (s ServiceImpl) IsFollower(ctx context.Context, follower, followed string) (bool, error) {
	const op = "follow.Service.IsFollower"

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.log.Error("failed to acquire connection", sl.Str("op", op))
		return false, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	_, err = q.Select(ctx, db.FollowSelectParams{
		Name:   follower,
		Name_2: followed,
	})
	if err != nil {
		s.log.Error("failed to determine whether user is follower", sl.Str("op", op))
		return false, err
	}

	return true, nil
}

func (s ServiceImpl) List(ctx context.Context, follower string) ([]FollowerListItem, error) {
	const op = "follow.Service.List"

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.log.Error("failed to acquire connection", sl.Str("op", op))
		return nil, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	list, err := q.SelectMany(ctx, follower)
	if err != nil {
		s.log.Error("failed to get following list", sl.Str("op", op))
		return nil, err
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
	const op = "follow.Service.ListExtended"

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.log.Error("failed to acquire connection", sl.Str("op", op))
		return nil, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	list, err := q.SelectManyExtended(ctx, follower)
	if err != nil {
		s.log.Error("failed to get extended following list", sl.Str("op", op))
		return nil, err
	}

	following := make([]FollowingListExtendedItem, len(list))
	for i, f := range list {
		following[i] = FollowingListExtendedItem{
			Name:     f.Username,
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
	const op = "follow.Service.Follow"

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.log.Error("failed to acquire connection", sl.Str("op", op))
		return err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	err = q.Insert(ctx, db.FollowInsertParams{
		Column1: pgtype.Text{String: follower, Valid: true},
		Column2: pgtype.Text{String: followed, Valid: true},
	})
	if err != nil {
		s.log.Error("failed to follow", sl.Str("op", op))
		return err
	}

	return nil
}

func (s ServiceImpl) Unfollow(ctx context.Context, unfollower, unfollowed string) error {
	const op = "follow.Service.Unfollow"

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.log.Error("failed to acquire connection", sl.Str("op", op))
		return err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	err = q.Delete(ctx, db.FollowDeleteParams{
		Name:   unfollower,
		Name_2: unfollowed,
	})
	if err != nil {
		s.log.Error("failed to unfollow", sl.Str("op", op))
		return err
	}

	return nil
}
