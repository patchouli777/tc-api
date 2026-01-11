package livestream

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// TODO: ttl 10min for livestreams
type cache struct {
	rdb *redis.Client
	// set of livestream ids sorted by viewers
	sorted *sortedIDStore
	// livestream store
	store *livestreamStore
	// map from username (channel) to livestream id
	userMap *userToIdStore
	// set of livestream ids
	ids *idStore
}

func newCache(rdb *redis.Client) *cache {
	sorted := sortedIDStore{rdb: rdb}
	store := livestreamStore{rdb: rdb}
	userMap := userToIdStore{rdb: rdb}
	ids := idStore{rdb: rdb}

	return &cache{rdb: rdb,
		sorted:  &sorted,
		store:   &store,
		userMap: &userMap,
		ids:     &ids}
}

func (r *cache) add(ctx context.Context, ls Livestream) error {
	lsIdStr := strconv.Itoa(int(ls.Id))

	cmds, err := r.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		r.userMap.addTx(ctx, p, ls.User.Name, lsIdStr)
		r.sorted.addTx(ctx, p, ls.Category.Link, ls.Viewers, lsIdStr)
		r.store.addTx(ctx, p, ls)
		r.ids.addTx(ctx, p, lsIdStr)
		return nil
	})

	if err != nil {
		return fmt.Errorf("pipeline failed: %w", err)
	}

	for i, cmd := range cmds {
		if cmd.Err() != nil {
			return fmt.Errorf("cmd %d failed: %w", i, cmd.Err())
		}
	}

	return nil
}

func (r *cache) get(ctx context.Context, username string) (*Livestream, error) {
	lsId, err := r.userMap.get(ctx, username)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errNotFound
		}

		return nil, fmt.Errorf("unable to find %s's livestream: %w", username, err)
	}

	ls, err := r.store.get(ctx, lsId)
	if err != nil {
		return nil, err
	}

	return ls, nil
}

func (r *cache) listAll(ctx context.Context) ([]Livestream, error) {
	ids, err := r.ids.getAll(ctx)
	if err != nil {
		return nil, err
	}

	livestreams, err := r.store.list(ctx, ids)
	if err != nil {
		return nil, err
	}

	return livestreams, nil
}

func (r *cache) list(ctx context.Context, category string, page, count int) ([]Livestream, error) {
	ids, err := r.sorted.get(ctx, category, page, count)
	if err != nil {
		return nil, err
	}

	res, err := r.store.list(ctx, ids)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *cache) update(ctx context.Context, lsId int32, title string, u User, c Category) (*Livestream, error) {
	lsIdStr := strconv.Itoa(int(lsId))

	cmds, err := r.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		r.userMap.addTx(ctx, p, u.Name, lsIdStr)
		r.store.updateFieldTx(ctx, p, lsIdStr, map[string]any{
			"user":     u,
			"category": c,
			"title":    title})
		// r.store.updateFieldTx(ctx, p, lsIdStr, "category", c)
		// r.store.updateFieldTx(ctx, p, lsIdStr, "title", title)
		r.ids.addTx(ctx, p, lsIdStr).Err()
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("pipeline failed: %w", err)
	}

	for i, cmd := range cmds {
		if cmd.Err() != nil {
			return nil, fmt.Errorf("cmd %d failed: %w", i, cmd.Err())
		}
	}

	livestream, err := r.store.get(ctx, lsIdStr)
	if err != nil {
		return nil, err
	}

	return livestream, nil
}

func (r *cache) updateThumbnail(ctx context.Context, user, thumbnail string) error {
	lsId, err := r.userMap.get(ctx, user)
	if err != nil {
		return nil
	}

	return r.store.updateThumbnail(ctx, lsId, thumbnail)
}

func (r *cache) updateViewers(ctx context.Context, user string, viewers int32) error {
	id, _ := r.userMap.get(ctx, user)
	ls, _ := r.store.get(ctx, id)
	cmds, err := r.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		r.sorted.addTx(ctx, p, ls.Category.Link, viewers, id)
		r.store.updateViewers(ctx, id, int(viewers))
		return nil
	})

	if err != nil {
		return fmt.Errorf("pipeline failed: %w", err) // EXEC-level error
	}

	for i, cmd := range cmds {
		if cmd.Err() != nil {
			return fmt.Errorf("cmd %d failed: %w", i, cmd.Err())
		}
	}

	return nil
}

func (r *cache) delete(ctx context.Context, username string) error {
	lsId, err := r.userMap.get(ctx, username)
	if err != nil {
		return err
	}

	if lsId == "" {
		return errAlreadyEnded
	}

	ls, err := r.store.get(ctx, lsId)
	if err != nil {
		return err
	}

	cmds, err := r.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		r.userMap.deleteTx(ctx, p, username)
		r.sorted.deleteTx(ctx, p, ls.Category.Link)
		r.store.deleteTx(ctx, p, lsId)
		r.ids.deleteTx(ctx, p, lsId)
		return nil
	})

	if err != nil {
		return fmt.Errorf("pipeline failed: %w", err)
	}

	for i, cmd := range cmds {
		if cmd.Err() != nil {
			return fmt.Errorf("cmd %d failed: %w", i, cmd.Err())
		}
	}

	return nil
}

func (r *cache) exists(ctx context.Context, username string) (bool, error) {
	id, err := r.userMap.get(ctx, username)
	if err != nil || id == "" {
		return false, err
	}

	return true, nil
}
