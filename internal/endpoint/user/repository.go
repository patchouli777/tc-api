package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{pool: pool}
}

func (r *RepositoryImpl) Get(ctx context.Context, username string) (*User, error) {
	return nil, errors.New("not implemented")
}

func (r *RepositoryImpl) Create(ctx context.Context, u UserCreate) error {
	return errors.New("not implemented")
}

func (r *RepositoryImpl) Update(ctx context.Context, u UserUpdate) error {
	return errors.New("not implemented")
}

func (r *RepositoryImpl) Delete(ctx context.Context, id int) error {
	return errors.New("not implemented")
}

func (r *RepositoryImpl) List(ctx context.Context, ul UserList) ([]User, error) {
	return nil, errors.New("not implemented")
}
