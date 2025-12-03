package user

import (
	"context"
	"log/slog"
	"main/internal/lib/sl"
)

type Repository interface {
	List(ctx context.Context, ul UserList) ([]User, error)
	Get(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, u UserCreate) error
	Update(ctx context.Context, u UserUpdate) error
	Delete(ctx context.Context, id int) error
}

type ServiceImpl struct {
	r   Repository
	log *slog.Logger
}

func NewService(log *slog.Logger, r Repository) *ServiceImpl {
	return &ServiceImpl{log: log, r: r}
}

func (s *ServiceImpl) List(ctx context.Context, ul UserList) ([]User, error) {
	const op = "user.Service.List"

	users, err := s.r.List(ctx, ul)
	if err != nil {
		s.log.Error(op, sl.Err(err))
		return nil, err
	}

	return users, nil
}

func (s *ServiceImpl) Get(ctx context.Context, username string) (*User, error) {
	const op = "user.Service.Get"

	user, err := s.r.Get(ctx, username)
	if err != nil {
		s.log.Error(op, sl.Err(err))
		return nil, err
	}

	return user, nil
}

func (s *ServiceImpl) Create(ctx context.Context, uc UserCreate) error {
	const op = "user.Service.Create"

	err := s.r.Create(ctx, uc)
	if err != nil {
		s.log.Error(op, sl.Err(err))
		return err
	}

	return nil
}

func (s *ServiceImpl) Update(ctx context.Context, uu UserUpdate) error {
	const op = "user.Service.Update"

	err := s.r.Update(ctx, uu)
	if err != nil {
		s.log.Error(op, sl.Err(err))
		return err
	}

	return nil
}

func (s *ServiceImpl) Delete(ctx context.Context, id int) error {
	const op = "user.Service.Delete"

	err := s.r.Delete(ctx, id)
	if err != nil {
		s.log.Error(op, sl.Err(err))
		return err
	}

	return nil
}
