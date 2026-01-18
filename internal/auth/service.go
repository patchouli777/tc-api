package auth

import (
	"context"
	"errors"
	ext "main/internal/external/auth"
	"main/internal/external/db"
)

type Client interface {
	GetPair(ctx context.Context, ui ext.UserInfo) (*ext.TokenPair, error)
}

type ServiceImpl struct {
	queries *QueriesAdapter
	ac      Client
}

func NewService(ac Client, adapter *QueriesAdapter) *ServiceImpl {
	return &ServiceImpl{ac: ac, queries: adapter}
}

func (s *ServiceImpl) SignIn(ctx context.Context, username, password string) (*TokenPair, error) {
	ui, err := s.queries.Select(ctx, db.AuthSelectUserParams{Name: username, Password: password})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, errWrongCredentials
		}

		return nil, err
	}

	pair, err := s.ac.GetPair(ctx, ext.UserInfo{Id: ui.ID, Username: username, Role: string(ui.AppRole)})
	if err != nil {
		return nil, err
	}

	return &TokenPair{Access: Token(pair.Access), Refresh: Token(pair.Refresh)}, nil
}

func (s *ServiceImpl) SignUp(ctx context.Context, email, username, password string) (*TokenPair, error) {
	err := s.queries.Insert(ctx, db.AuthInsertUserParams{Name: username, Password: password})
	if err != nil {
		return nil, err
	}

	pair, err := s.ac.GetPair(ctx, ext.UserInfo{Username: username, Role: "user"})
	if err != nil {
		return nil, err
	}

	return &TokenPair{Access: Token(pair.Access), Refresh: Token(pair.Refresh)}, nil
}
