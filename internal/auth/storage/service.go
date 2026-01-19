package storage

import (
	"context"
	"errors"
	d "main/internal/auth/domain"
	ext "main/internal/external/auth"
	"main/internal/external/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	GetPair(ctx context.Context, ui ext.UserInfo) (*ext.TokenPair, error)
}

type ServiceImpl struct {
	pool *pgxpool.Pool
	ac   Client
}

func NewService(ac Client, pool *pgxpool.Pool) *ServiceImpl {
	return &ServiceImpl{ac: ac, pool: pool}
}

func (s *ServiceImpl) SignIn(ctx context.Context, username, password string) (*d.TokenPair, error) {
	q := queriesAdapter{queries: db.New(s.pool)}
	ui, err := q.Select(ctx, db.AuthSelectUserParams{Name: username, Password: password})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, d.ErrWrongCredentials
		}

		return nil, err
	}

	pair, err := s.ac.GetPair(ctx, ext.UserInfo{Id: ui.ID, Username: username, Role: string(ui.AppRole)})
	if err != nil {
		return nil, err
	}

	return &d.TokenPair{Access: d.Token(pair.Access), Refresh: d.Token(pair.Refresh)}, nil
}

func (s *ServiceImpl) SignUp(ctx context.Context, email, username, password string) (*d.TokenPair, error) {
	q := queriesAdapter{queries: db.New(s.pool)}
	err := q.Insert(ctx, db.AuthInsertUserParams{Name: username, Password: password})
	if err != nil {
		return nil, err
	}

	pair, err := s.ac.GetPair(ctx, ext.UserInfo{Username: username, Role: "user"})
	if err != nil {
		return nil, err
	}

	return &d.TokenPair{Access: d.Token(pair.Access), Refresh: d.Token(pair.Refresh)}, nil
}
