package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/db"
	"main/internal/lib/sl"
)

// TODO: refresh, access unused
type GRPCCLient interface {
	GetRefresh(ctx context.Context, ui UserInfo) (*Token, error)
	GetAccess(ctx context.Context, ui UserInfo) (*Token, error)
	GetPair(ctx context.Context, ui UserInfo) (*TokenPair, error)
}

type ServiceImpl struct {
	log     *slog.Logger
	adapter *QueriesAdapter
	grpc    GRPCCLient
}

func NewService(log *slog.Logger, grpc GRPCCLient, adapter *QueriesAdapter) *ServiceImpl {
	return &ServiceImpl{log: log, grpc: grpc, adapter: adapter}
}

func (s *ServiceImpl) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	const op = "auth.ServiceImpl.Login"

	ui, err := s.adapter.Select(ctx, db.AuthSelectUserParams{Name: username, Password: password})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			s.log.Error("error finding user", sl.Err(err))
			return nil, ErrWrongCredentials
		}

		s.log.Error("error finding user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pair, err := s.grpc.GetPair(ctx, UserInfo{Id: ui.ID, Username: username, Role: string(ui.AppRole)})
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (s *ServiceImpl) Register(ctx context.Context, email, username, password string) (*TokenPair, error) {
	err := s.adapter.Insert(ctx, db.AuthInsertUserParams{Name: username, Password: password})
	if err != nil {
		if errors.Is(err, db.ErrDuplicateKey) {
			return nil, ErrUserAlreadyExists
		}

		s.log.Error("error register", sl.Err(err))
		return nil, err
	}

	pair, err := s.grpc.GetPair(ctx, UserInfo{Username: username, Role: "user"})
	if err != nil {
		return nil, err
	}

	return pair, nil
}
