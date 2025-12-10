package channel

import (
	"context"
	"errors"
	"log/slog"
)

type ServiceImpl struct {
	log     *slog.Logger
	queries *QueriesAdapter
}

func NewService(log *slog.Logger, q *QueriesAdapter) *ServiceImpl {
	return &ServiceImpl{log: log, queries: q}
}

func (s *ServiceImpl) Get(ctx context.Context, chann string) (*Channel, error) {
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
		Links:           channel.Links,
		Tags:            channel.Tags,
	}, nil
}

func (s *ServiceImpl) Update(ctx context.Context, upd ChannelUpdate) error {
	return errors.New("not implemented")
}
