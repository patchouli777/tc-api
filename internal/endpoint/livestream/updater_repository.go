package livestream

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type UpdaterRepositoryImpl struct {
	rdb *redis.Client
}

func NewUpdaterRepo(rdb *redis.Client) *UpdaterRepositoryImpl {
	return &UpdaterRepositoryImpl{rdb: rdb}
}

func (r *UpdaterRepositoryImpl) GetLivestreamViewers(ctx context.Context, channel string) (*int, error) {
	key := r.channelToViewersByInstanceKey(channel)
	res, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	viewers := 0
	for _, v := range res {
		views, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		viewers += views
	}

	return &viewers, nil
}

func (r *UpdaterRepositoryImpl) GetUpdatableLivestream(ctx context.Context) (string, error) {
	key := r.updatableLivestreamsKey()
	return r.rdb.SRandMember(ctx, key).Result()
}

func (r *UpdaterRepositoryImpl) DeleteUpdatableLivestream(ctx context.Context, channel string) error {
	key := r.updatableLivestreamsKey()
	return r.rdb.SRem(ctx, key, channel).Err()
}

func (r *UpdaterRepositoryImpl) channelToViewersByInstanceKey(channel string) string {
	return fmt.Sprintf("viewers:%s", channel)
}

func (r *UpdaterRepositoryImpl) updatableLivestreamsKey() string {
	return "update:livestream"
}
