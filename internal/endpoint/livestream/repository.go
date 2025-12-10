package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"main/internal/db"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type LivestreamId int

func (m LivestreamId) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// TODO: sync redis and pg

type RepositoryImpl struct {
	rdb  *redis.Client
	pool *pgxpool.Pool
	// set of livestream ids sorted by viewers
	sorted *sortedIDStore
	// livestream store
	store *livestreamStore
	// map from username to livestream id
	userMap *userToIdStore
	// set of livestream ids
	ids *idStore
}

func NewRepo(rdb *redis.Client, pool *pgxpool.Pool) *RepositoryImpl {
	sorted := sortedIDStore{rdb: rdb}
	store := livestreamStore{rdb: rdb}
	userMap := userToIdStore{rdb: rdb}
	ids := idStore{rdb: rdb}

	return &RepositoryImpl{rdb: rdb,
		pool:    pool,
		sorted:  &sorted,
		store:   &store,
		userMap: &userMap,
		ids:     &ids}
}

func (r *RepositoryImpl) Create(ctx context.Context, cr LivestreamCreate) (*Livestream, error) {
	id, err := r.userMap.Get(ctx, cr.Username)

	// TODO: sentinel error
	if id != nil {
		return nil, fmt.Errorf("трансляция пользователя %s уже начата: %+v", cr.Username, id)
	}

	if err != nil && err != redis.Nil {
		return nil, err
	}

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	ins, err := q.Insert(ctx, db.LivestreamInsertParams{
		Name:  cr.Username,
		Link:  cr.CategoryLink,
		Title: pgtype.Text{String: cr.Title, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	ls := Livestream{
		Id:        ins.LivestreamID,
		Title:     cr.Title,
		StartedAt: int64(ins.StartedAt.Int32),
		User: User{
			Name:   ins.UserName,
			Avatar: ins.UserAvatar.String,
		},
		Category: Category{
			Name: ins.CategoryName,
			Link: ins.CategoryLink,
		},
		Viewers:   0,
		Thumbnail: "",
	}

	err = r.addLivestream(ctx, ls)
	if err != nil {
		return nil, err
	}

	return &ls, nil
}

func (r *RepositoryImpl) Get(ctx context.Context, username string) (*Livestream, error) {
	lsId, err := r.userMap.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти трансляцию пользователя %s: %v", username, err)
	}

	ls, err := r.store.Get(ctx, *lsId)
	if err != nil {
		return nil, err
	}

	return ls, nil
}

// TODO: impl
func (r *RepositoryImpl) ListById(ctx context.Context, categoryId string, page, count int) ([]Livestream, error) {
	// categoryLink, err := r.stringLinkGet(ctx, categoryId)
	// if err != nil {
	// 	return nil, err
	// }

	// ids, err := r.sorted.Get(ctx, *categoryLink, page, count)
	// if err != nil {
	// 	return nil, err
	// }

	// return r.store.List(ctx, ids)
	return nil, errors.New("not implemented")
}

func (r *RepositoryImpl) ListAll(ctx context.Context) ([]Livestream, error) {
	ids, err := r.ids.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	livestreams, err := r.store.List(ctx, ids)
	if err != nil {
		return nil, err
	}

	return livestreams, nil
}

func (r *RepositoryImpl) List(ctx context.Context, category string, page, count int) ([]Livestream, error) {
	ids, err := r.sorted.Get(ctx, category, page, count)
	if err != nil {
		return nil, err
	}

	res, err := r.store.List(ctx, ids)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *RepositoryImpl) UpdateViewers(ctx context.Context, user string, viewers int32) error {
	id, err := r.userMap.Get(ctx, user)
	if err != nil {
		return err
	}

	ls, err := r.store.Get(ctx, *id)
	if err != nil {
		return err
	}

	if err = r.sorted.Add(ctx, ls.Category.Link, viewers, *id).Err(); err != nil {
		return err
	}

	return r.store.UpdateViewers(ctx, *id, int(viewers))
}

func (r *RepositoryImpl) UpdateThumbnail(ctx context.Context, user, thumbnail string) error {
	lsId, err := r.userMap.Get(ctx, user)
	if err != nil {
		return nil
	}

	return r.store.UpdateThumbnail(ctx, *lsId, thumbnail)
}

func (r *RepositoryImpl) Update(ctx context.Context, cur *Livestream, upd LivestreamUpdate) (*Livestream, error) {
	if upd.CategoryLink != nil {
		cur.Category.Link = *upd.CategoryLink
	}

	if upd.Viewers != nil {
		cur.Viewers = *upd.Viewers
	}

	if upd.Thumbnail != nil {
		cur.Thumbnail = *upd.Thumbnail
	}

	if upd.Title != nil {
		cur.Title = *upd.Title
	}

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	updated, err := q.Update(ctx, db.LivestreamUpdateParams{
		Title:   pgtype.Text{String: cur.Title, Valid: true},
		Viewers: pgtype.Int4{Int32: cur.Viewers, Valid: true},
		Link:    cur.Category.Link,
		Name:    cur.User.Name,
	})
	if err != nil {
		return nil, err
	}

	cur.Category.Link = updated.CategoryLink
	cur.Category.Name = updated.CategoryName
	cur.User.Name = updated.UserName
	cur.User.Avatar = updated.UserAvatar.String

	err = r.addLivestream(ctx, *cur)
	if err != nil {
		return nil, err
	}

	return cur, nil
}

// TODO: sentinel error to return 200 on multiple deletes
func (r *RepositoryImpl) Delete(ctx context.Context, username string) (bool, error) {
	lsId, err := r.userMap.Get(ctx, username)
	if err != nil {
		return false, err
	}

	if lsId == nil {
		return false, fmt.Errorf("трансляция уже оффлайн")
	}

	ls, err := r.store.Get(ctx, *lsId)
	if err != nil {
		return false, err
	}

	if err = r.userMap.Delete(ctx, username).Err(); err != nil {
		return false, err
	}

	if err = r.sorted.Delete(ctx, ls.Category.Link).Err(); err != nil {
		return false, err
	}

	if err = r.store.Delete(ctx, *lsId).Err(); err != nil {
		return false, err
	}

	if err = r.ids.Delete(ctx, *lsId); err != nil {
		return false, err
	}

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	err = q.Delete(ctx, username)
	if err != nil {
		return false, err
	}

	return true, nil
}

// TODO: transactions/pipeline
func (r *RepositoryImpl) addLivestream(ctx context.Context, ls Livestream) error {
	lsIdStr := strconv.Itoa(int(ls.Id))

	res, err := r.userMap.Add(ctx, ls.User.Name, lsIdStr)
	if err != nil {
		return fmt.Errorf("не удалось сохранить данные о трансляции в строку: %v", err)
	}
	if res.Err() != nil {
		return err
	}

	if err = r.sorted.Add(ctx, ls.Category.Link, ls.Viewers, lsIdStr).Err(); err != nil {
		return fmt.Errorf("не удалось сохранить id трансляции в сортированное множество: %v", err)
	}

	if err = r.store.Add(ctx, ls).Err(); err != nil {
		return fmt.Errorf("не удалось сохранить данные о трансляции в множество: %v", err)
	}

	if err = r.ids.Add(ctx, lsIdStr); err != nil {
		return fmt.Errorf("не удалось сохранить id Трансляции в множество: %v", err)
	}

	return nil
}
