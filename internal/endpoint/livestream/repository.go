package livestream

import (
	"context"
	"main/internal/external/db"

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

func (r *RepositoryImpl) Get(ctx context.Context, username string) (*Livestream, error) {
	return r.cache.get(ctx, username)
}

func (r *RepositoryImpl) ListAll(ctx context.Context) ([]Livestream, error) {
	return r.cache.listAll(ctx)
}

func (r *RepositoryImpl) List(ctx context.Context, category string, page, count int) ([]Livestream, error) {
	return r.cache.list(ctx, category, page, count)
}

func (r *RepositoryImpl) Create(ctx context.Context, cr LivestreamCreate) (*Livestream, error) {
	exists, err := r.cache.exists(ctx, cr.Username)

	if exists {
		return nil, errAlreadyStarted
	}

	if err != nil && err != redis.Nil {
		return nil, err
	}

	q := QueriesAdapter{queries: db.New(r.pool)}

	ins, err := q.Insert(ctx, db.LivestreamInsertParams{
		Name:  cr.Username,
		Link:  cr.Category,
		Title: cr.Title,
	})
	if err != nil {
		return nil, err
	}

	ls := Livestream{
		Id:           ins.LivestreamID,
		Title:        cr.Title,
		StartedAt:    int64(ins.StartedAt.Time.Second()),
		UserName:     ins.UserName,
		UserAvatar:   ins.UserAvatar.String,
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

func (r *RepositoryImpl) UpdateViewers(ctx context.Context, user string, viewers int32) error {
	// TODO: batch updates
	// go func() {
	// 	q := QueriesAdapter{queries: db.New(r.pool)}

	// 	err := q.UpdateViewers(ctx, db.LivestreamUpdateViewersParams{
	// 		Viewers: viewers,
	// 		Name:    user,
	// 	})
	// 	if err != nil {
	// 		slog.Error("unable to update viewers", sl.Err(err))
	// 	}
	// }()

	return r.cache.updateViewers(ctx, user, viewers)
}

func (r *RepositoryImpl) Update(ctx context.Context, channel string, upd LivestreamUpdate) (*Livestream, error) {
	q := QueriesAdapter{queries: db.New(r.pool)}
	// NOTE: for some reason sqlc generates types like that
	// TODO: fix someday?
	// TODO: pass user id and not username because username can be changed in realtime
	updated, err := q.Update(ctx, db.LivestreamUpdateParams{
		Title: pgtype.Text{
			String: upd.Title.Value,
			Valid:  upd.Title.Explicit && !upd.Title.IsNull},
		TitleDoUpdate: pgtype.Bool{
			Bool:  upd.Title.Explicit && !upd.Title.IsNull,
			Valid: true},

		IDCategory: pgtype.Int4{
			Int32: int32(upd.CategoryId.Value),
			Valid: upd.CategoryId.Explicit && !upd.CategoryId.IsNull,
		},
		IDCategoryDoUpdate: pgtype.Bool{
			Bool:  upd.CategoryId.Explicit && !upd.CategoryId.IsNull,
			Valid: true,
		},

		Username: pgtype.Text{
			String: channel,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, err
	}

	return r.cache.update(ctx,
		updated.LivestreamID,
		updated.Title,
		User{
			Name:   updated.UserName,
			Avatar: updated.UserAvatar.String,
		}, Category{
			Name: updated.CategoryName,
			Link: updated.CategoryLink,
		})
}

func (r *RepositoryImpl) UpdateThumbnail(ctx context.Context, user, thumbnail string) error {
	return r.cache.updateThumbnail(ctx, user, thumbnail)
}

func (r *RepositoryImpl) Delete(ctx context.Context, username string) error {
	err := r.cache.delete(ctx, username)
	if err != nil {
		return err
	}

	q := QueriesAdapter{queries: db.New(r.pool)}
	return q.Delete(ctx, username)
}
