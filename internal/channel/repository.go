package channel

import (
	"context"
	"errors"
	"log/slog"
)

type RepositoryImpl struct {
	log     *slog.Logger
	queries *QueriesAdapter
}

func NewRepository(log *slog.Logger, q *QueriesAdapter) *RepositoryImpl {
	return &RepositoryImpl{log: log, queries: q}
}

func (s *RepositoryImpl) Get(ctx context.Context, chann string) (*Channel, error) {
	channel, err := s.queries.Select(ctx, chann)
	if err != nil {
		return nil, err
	}

	return &Channel{
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

func (s *RepositoryImpl) Update(ctx context.Context, upd ChannelUpdate) error {
	return errors.New("not implemented")
}
