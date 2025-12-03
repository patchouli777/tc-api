package presence

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"

	"github.com/redis/go-redis/v9"
)

type ViewerStore struct {
	rdb *redis.Client
}

func NewViewerStore(rdb *redis.Client) *ViewerStore {
	return &ViewerStore{rdb: rdb}
}

func (r *ViewerStore) AddInstanceStat(ctx context.Context,
	channel, instanceID string,
	viewers int,
	thumbnail string) error {
	fmt.Printf("adding channel %s from instance %s\n", channel, instanceID)
	err := r.channelToViewersByInstanceAdd(ctx, channel, instanceID, viewers)
	if err != nil {
		slog.Error("unable to add viewers", sl.Err(err), sl.Str("instanceID", instanceID))
		return err
	}

	err = r.updatableLivestreamsAdd(ctx, channel)
	if err != nil {
		slog.Error("unable to mark channel as updatable", sl.Err(err), sl.Str("instanceID", instanceID))
		return err
	}

	return nil
}

// TODO: publish channel name in pub/sub
// updater subscribes to pubsub, reads channels that needs to be updated and holds
// update timer for them and perform updates accordingly

func (r *ViewerStore) PublishChannel(ctx context.Context, channel string) error {
	return r.rdb.Publish(ctx, "updatable-channel", channel).Err()
}

func (r *ViewerStore) channelToViewersByInstanceAdd(ctx context.Context, channel, instanceID string, viewers int) error {
	key := fmt.Sprintf("viewers:%s", channel)
	return r.rdb.HSet(ctx, key, instanceID, viewers).Err()
}

func (r *ViewerStore) channelToThumbnailAdd(ctx context.Context, channel, thumbnail string) error {
	key := fmt.Sprintf("viewers:%s", channel)
	return r.rdb.Set(ctx, key, thumbnail, 0).Err()
}

func (r *ViewerStore) updatableLivestreamsAdd(ctx context.Context, channel string) error {
	key := "update:livestream"
	return r.rdb.SAdd(ctx, key, channel).Err()
}
