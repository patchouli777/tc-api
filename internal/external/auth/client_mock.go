package auth

import (
	"context"
	"main/internal/auth"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ClientMock struct{}

func (s *ClientMock) GetRefresh(ctx context.Context, ui UserInfo) (*Token, error) {
	now := time.Now()
	refreshExp := now.Add(refreshExpirationTime)

	refresh, err := generateJWT(ui, refreshExp)
	if err != nil {
		return nil, err
	}

	tok := Token(refresh)

	return &tok, nil
}

func (s *ClientMock) GetAccess(ctx context.Context, ui UserInfo) (*Token, error) {
	now := time.Now()
	accessExp := now.Add(accessExpirationTime)

	access, err := generateJWT(ui, accessExp)
	if err != nil {
		return nil, err
	}

	tok := Token(access)

	return &tok, nil
}

func (s *ClientMock) GetPair(ctx context.Context, ui UserInfo) (*TokenPair, error) {
	return createPair(ui)
}

func generateJWT(ui UserInfo, expirationTime time.Time) (string, error) {
	claims := &auth.Claims{
		Id:       ui.Id,
		Username: ui.Username,
		Role:     ui.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(auth.JWTKey)
}

func createPair(ui UserInfo) (*TokenPair, error) {
	now := time.Now()
	refreshExp := now.Add(refreshExpirationTime)
	accessExp := now.Add(accessExpirationTime)

	refresh, err := generateJWT(ui, refreshExp)
	if err != nil {
		return nil, err
	}

	access, err := generateJWT(ui, accessExp)
	if err != nil {
		return nil, err
	}

	return &TokenPair{Access: Token(access), Refresh: Token(refresh)}, nil
}

const refreshExpirationTime = time.Duration(14400) * time.Minute
const accessExpirationTime = time.Duration(120) * time.Minute
