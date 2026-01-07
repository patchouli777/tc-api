package auth

import (
	"context"
	"errors"
	"main/internal/db"
)

// TODO: refresh, access unused
type GRPCCLient interface {
	GetRefresh(ctx context.Context, ui UserInfo) (*Token, error)
	GetAccess(ctx context.Context, ui UserInfo) (*Token, error)
	GetPair(ctx context.Context, ui UserInfo) (*TokenPair, error)
}

type ServiceImpl struct {
	queries *QueriesAdapter
	grpc    GRPCCLient
}

func NewService(grpc GRPCCLient, adapter *QueriesAdapter) *ServiceImpl {
	return &ServiceImpl{grpc: grpc, queries: adapter}
}

// TODO: grpc client stuff
func (s *ServiceImpl) SignIn(ctx context.Context, username, password string) (*TokenPair, error) {
	ui, err := s.queries.Select(ctx, db.AuthSelectUserParams{Name: username, Password: password})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, errWrongCredentials
		}

		return nil, err
	}

	pair, err := s.grpc.GetPair(ctx, UserInfo{Id: ui.ID, Username: username, Role: string(ui.AppRole)})
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (s *ServiceImpl) SignUp(ctx context.Context, email, username, password string) (*TokenPair, error) {
	err := s.queries.Insert(ctx, db.AuthInsertUserParams{Name: username, Password: password})
	if err != nil {
		if errors.Is(err, db.ErrDuplicateKey) {
			return nil, errUserAlreadyExists
		}

		return nil, err
	}

	pair, err := s.grpc.GetPair(ctx, UserInfo{Username: username, Role: "user"})
	if err != nil {
		return nil, err
	}

	return pair, nil
}
