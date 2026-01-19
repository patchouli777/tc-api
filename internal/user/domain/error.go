package user

import "errors"

var (
	ErrNotFound         = errors.New("user not found")
	ErrAlreadyExists    = errors.New("user already exists")
	ErrWeakPassword     = errors.New("password is too weak")
	ErrUsernameRequired = errors.New("username required")
	ErrPasswordRequired = errors.New("password required")
)
