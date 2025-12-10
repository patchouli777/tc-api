package livestream

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"
)

type LivestreamSearch struct {
	CategoryId string
	Category   string
	Page       int
	Count      int
}

type Repository interface {
	Create(ctx context.Context, cr LivestreamCreate) (*Livestream, error)
	Update(ctx context.Context, cur *Livestream, upd LivestreamUpdate) (*Livestream, error)
	UpdateViewers(ctx context.Context, user string, viewers int32) error
	UpdateThumbnail(ctx context.Context, user, thumbnail string) error
	Delete(ctx context.Context, username string) (bool, error)
	Get(ctx context.Context, username string) (*Livestream, error)
	List(ctx context.Context, category string, page, count int) ([]Livestream, error)
	ListAll(ctx context.Context) ([]Livestream, error)
	ListById(ctx context.Context, categoryId string, page, count int) ([]Livestream, error)
}

type StreamServerAdapter interface {
	Start(ctx context.Context, username string) error
	End(ctx context.Context, username string) error
	List(ctx context.Context) (*StreamServerResponse, error)
	Update(ctx context.Context)
}

type ServiceImpl struct {
	log     *slog.Logger
	repo    Repository
	adapter StreamServerAdapter
}

func NewService(log *slog.Logger, repo Repository, adapter StreamServerAdapter) *ServiceImpl {
	return &ServiceImpl{log: log, repo: repo, adapter: adapter}
}

func (s ServiceImpl) Start(ctx context.Context, categoryLink, title, username string) (*Livestream, error) {
	err := s.adapter.Start(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("unable to start livestream: %w", err)
	}

	ls, err := s.repo.Create(ctx, LivestreamCreate{
		Title:        title,
		CategoryLink: categoryLink,
		Username:     username,
	})
	if err != nil {
		s.adapter.End(ctx, username)
		return nil, fmt.Errorf("unable to cache livestream: %w", err)
	}

	return ls, nil
}

func (s ServiceImpl) Get(ctx context.Context, username string) (*Livestream, error) {
	ls, err := s.repo.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("unable to find livestream: %w", err)
	}

	return ls, nil
}

func (s ServiceImpl) List(ctx context.Context, search LivestreamSearch) ([]Livestream, error) {
	if search.Page < 1 {
		search.Page = 1
	}

	if search.Count < 1 {
		search.Count = 10
	}

	if search.CategoryId != "" {
		list, err := s.repo.ListById(ctx, search.CategoryId, search.Page, search.Count)
		if err != nil {
			return nil, fmt.Errorf("unable to get list of livestreams: %v", err)
		}
		return list, nil
	}

	list, err := s.repo.List(ctx, search.Category, search.Page, search.Count)

	if err != nil {
		return nil, fmt.Errorf("unable to get list of livestreams: %v", err)
	}
	return list, nil
}

func (s ServiceImpl) UpdateViewers(ctx context.Context, user string, viewers int32) error {
	err := s.repo.UpdateViewers(ctx, user, viewers)
	if err != nil {
		return err
	}

	return nil
}

func (s ServiceImpl) Update(ctx context.Context, user string, upd LivestreamUpdate) (bool, error) {
	current, err := s.repo.Get(ctx, user)
	if err != nil {
		return false, fmt.Errorf("unable to find livestream: %v", err)
	}

	_, err = s.repo.Update(ctx, current, upd)
	if err != nil {
		return false, fmt.Errorf("unable to update livestream: %v", err)
	}

	return true, nil
}

// TODO: 200 on multiple stop
func (s ServiceImpl) Stop(ctx context.Context, u string) (bool, error) {
	err := s.adapter.End(ctx, u)
	if err != nil {
		// return false, err
		s.log.Error("unable to end livestream", sl.Err(err))
	}

	res, err := s.repo.Delete(ctx, u)
	if err != nil {
		return false, fmt.Errorf("unable to delete livestream from cache: %v", err)
	}

	if !res {
		return false, fmt.Errorf("unable to find livestream")
	}

	return true, nil
}
