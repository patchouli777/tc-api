package auth

import "errors"

var (
	ErrTokenNotFound     = errors.New("token not found")
	ErrWrongCredentials  = errors.New("wrong credentials")
	ErrUserAlreadyExists = errors.New("user already exists")
)
