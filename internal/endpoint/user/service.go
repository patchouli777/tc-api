package user

import (
	"context"
	"fmt"
)

type Repository interface {
	List(ctx context.Context, ul UserList) ([]User, error)
	Get(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, u UserCreate) error
	Update(ctx context.Context, u UserUpdate) error
	Delete(ctx context.Context, id int) error
}

type ServiceImpl struct {
	r Repository
}

func NewService(r Repository) *ServiceImpl {
	return &ServiceImpl{r: r}
}

func (s *ServiceImpl) List(ctx context.Context, ul UserList) ([]User, error) {
	users, err := s.r.List(ctx, ul)
	if err != nil {
		return nil, fmt.Errorf("unable to list users: %w: ", err)
	}

	return users, nil
}

func (s *ServiceImpl) Get(ctx context.Context, username string) (*User, error) {
	user, err := s.r.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("unable to get user: %w", err)
	}

	return user, nil
}

func (s *ServiceImpl) Create(ctx context.Context, uc UserCreate) error {
	err := s.r.Create(ctx, uc)
	if err != nil {
		return fmt.Errorf("unable to create user: %w", err)
	}

	return nil
}

func (s *ServiceImpl) Update(ctx context.Context, uu UserUpdate) error {
	err := s.r.Update(ctx, uu)
	if err != nil {
		return fmt.Errorf("unable to update user: %w", err)
	}

	return nil
}

func (s *ServiceImpl) Delete(ctx context.Context, id int) error {
	err := s.r.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}

	return nil
}
