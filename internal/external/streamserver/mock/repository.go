package mock

import (
	"context"
	"errors"
	"slices"
)

type repository struct {
	livestreams []livestream
}

func newRepository() *repository {
	return &repository{livestreams: make([]livestream, 0)}
}

func (u *repository) List(ctx context.Context) ([]livestream, error) {
	return u.livestreams, nil
}

func (u *repository) Get(ctx context.Context, id string) (livestream, error) {
	for _, ls := range u.livestreams {
		if ls.channel == id {
			return ls, nil
		}
	}
	return livestream{}, errors.New("not found")
}

func (u *repository) Start(ctx context.Context, username string) error {
	for _, ls := range u.livestreams {
		if ls.channel == username {
			return nil
		}
	}

	u.livestreams = append(u.livestreams, livestream{channel: username, viewers: 0})
	return nil
}

func (u *repository) End(ctx context.Context, username string) error {
	for i, ls := range u.livestreams {
		if username != ls.channel {
			continue
		}

		u.livestreams = slices.Delete(u.livestreams, i, i+1)
		break
	}

	return nil
}
