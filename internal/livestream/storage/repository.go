package storage

import (
	"context"
	"fmt"
	"twitchy-api/internal/external/db"
	d "twitchy-api/internal/livestream/domain"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// TODO: sync redis and pg
// cleanup if updater can't get livestream data
// TODO: add to tc_livestream_history upon livestream end
type RepositoryImpl struct {
	pool *pgxpool.Pool
	// cache is actually primary database for livestreams
	cache *cache
}

func NewRepo(rdb *redis.Client, pool *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{
		cache: newCache(rdb),
		pool:  pool,
	}
}

func (r *RepositoryImpl) Get(ctx context.Context, id int) (*d.Livestream, error) {
	return r.cache.get(ctx, id)
}

func (r *RepositoryImpl) GetByUsername(ctx context.Context, username string) (*d.Livestream, error) {
	id, err := r.cache.userMap.get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%s's livestream not found: %v", username, err)
	}

	return r.cache.get(ctx, id)
}

func (r *RepositoryImpl) List(ctx context.Context, s d.LivestreamSearch) ([]d.Livestream, error) {
	return r.cache.list(ctx, s.Category, s.Page, s.Count)
}

func (r *RepositoryImpl) Create(ctx context.Context, cr d.LivestreamCreate) (*d.Livestream, error) {
	exists, err := r.cache.existsByUsername(ctx, cr.Username)

	if exists {
		return nil, d.ErrAlreadyStarted
	}

	if err != nil && err != redis.Nil {
		return nil, err
	}

	q := queriesAdapter{queries: db.New(r.pool)}

	ins, err := q.Insert(ctx, cr.Username)
	if err != nil {
		return nil, err
	}

	ls := d.Livestream{
		Id:           int(ins.LivestreamID),
		Title:        ins.Title.String,
		StartedAt:    ins.StartedAt.Time.Second(),
		UserId:       int(ins.UserID),
		UserName:     ins.UserName,
		UserPfp:      ins.UserPfp.String,
		CategoryId:   int(ins.CategoryID),
		CategoryName: ins.CategoryName,
		CategoryLink: ins.CategoryLink,
	}

	// TODO: insert success into cache failure -> big bad
	err = r.cache.add(ctx, ls)
	if err != nil {
		return nil, err
	}

	return &ls, nil
}

func (r *RepositoryImpl) UpdateViewers(ctx context.Context, id int, viewers int) error {
	// TODO: batch updates
	// go func() {
	// 	q := queriesAdapter{queries: db.New(r.pool)}

	// 	err := q.UpdateViewers(ctx, db.LivestreamUpdateViewersParams{
	// 		Viewers: viewers,
	// 		Name:    user,
	// 	})
	// 	if err != nil {
	// 		slog.Error("unable to update viewers", sl.Err(err))
	// 	}
	// }()

	return r.cache.updateViewers(ctx, id, viewers)
}

func (r *RepositoryImpl) Update(ctx context.Context, id int, upd d.LivestreamUpdate) (*d.Livestream, error) {
	q := queriesAdapter{queries: db.New(r.pool)}
	updated, err := q.Update(ctx, db.LivestreamUpdateParams{
		ID: int32(id),

		Title: pgtype.Text{String: upd.Title.Value,
			Valid: upd.Title.Explicit && !upd.Title.IsNull},
		TitleDoUpdate: upd.Title.Explicit && !upd.Title.IsNull,

		IDCategory:         pgtype.Int4{Int32: int32(upd.CategoryId.Value), Valid: true},
		IDCategoryDoUpdate: upd.CategoryId.Explicit && !upd.CategoryId.IsNull,
	})
	if err != nil {
		return nil, err
	}

	return r.cache.update(ctx,
		int(updated.LivestreamID),
		updated.Title.String,
		d.User{
			Name: updated.UserName,
			Pfp:  updated.UserPfp.String,
		}, d.Category{
			Name: updated.CategoryName,
			Link: updated.CategoryLink,
		})
}

func (r *RepositoryImpl) UpdateThumbnail(ctx context.Context, id int, thumbnail string) error {
	return r.cache.updateThumbnail(ctx, id, thumbnail)
}

func (r *RepositoryImpl) Delete(ctx context.Context, id int) error {
	err := r.cache.delete(ctx, id)
	if err != nil {
		return err
	}

	q := queriesAdapter{queries: db.New(r.pool)}
	return q.Delete(ctx, id)
}
