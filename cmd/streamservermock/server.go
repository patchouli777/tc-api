package main

import (
	"context"
	"errors"
	"log/slog"
	"slices"
)

type StreamServerMock struct {
	livestreams []livestream
	log         *slog.Logger
}

func NewStreamServerMock(log *slog.Logger) *StreamServerMock {
	return &StreamServerMock{log: log,
		livestreams: []livestream{}}
}

func (u *StreamServerMock) List(ctx context.Context) ([]livestream, error) {
	return u.livestreams, nil
}

func (u *StreamServerMock) Get(ctx context.Context, id string) (*livestream, error) {
	for _, ls := range u.livestreams {
		if ls.channel == id {
			copy := ls
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (u *StreamServerMock) Start(ctx context.Context, username string) error {
	u.livestreams = append(u.livestreams, livestream{channel: username, viewers: 0})
	return nil
}

func (u *StreamServerMock) End(ctx context.Context, username string) error {
	for i, ls := range u.livestreams {
		if username != ls.channel {
			continue
		}

		u.livestreams = slices.Delete(u.livestreams, i, i+1)
		break
	}

	return nil
}
