package channel

import (
	"context"
	"main/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChannelServiceImpl struct {
	pool *pgxpool.Pool
}

func (s ChannelServiceImpl) Get(ctx context.Context, chann string) (*Channel, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := QueriesAdapter{queries: db.New(conn)}
	channel, err := q.Select(ctx, chann)
	if err != nil {
		return nil, err
	}

	return &Channel{
		Name:            channel.Name,
		IsBanned:        channel.IsBanned.Bool,
		IsPartner:       channel.IsPartner.Bool,
		FirstLivestream: channel.FirstLivestream.Time,
		LastLivestream:  channel.LastLivestream.Time,
		Description:     channel.Description.String,
		Links:           channel.Links,
		Tags:            channel.Tags,
	}, nil
}
