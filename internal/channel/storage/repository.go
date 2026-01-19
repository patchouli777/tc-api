package storage

import (
	"context"
	"errors"

	d "main/internal/channel/domain"
	"main/internal/external/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{pool: pool}
}

func (s *RepositoryImpl) Get(ctx context.Context, chann string) (*d.Channel, error) {
	q := queriesAdapter{queries: db.New(s.pool)}

	channel, err := q.Select(ctx, chann)
	if err != nil {
		return nil, err
	}

	return &d.Channel{
		Name:            channel.Name,
		Background:      channel.Background,
		IsBanned:        channel.IsBanned.Bool,
		IsPartner:       channel.IsPartner.Bool,
		FirstLivestream: channel.FirstLivestream.Time,
		LastLivestream:  channel.LastLivestream.Time,
		Description:     channel.Description.String,
		// Links:           channel.Links,
		// Tags:            channel.Tags,
	}, nil
}

func (s *RepositoryImpl) Update(ctx context.Context, upd d.ChannelUpdate) error {
	return errors.New("not implemented")
}
