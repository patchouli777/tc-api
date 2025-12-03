package user

import "errors"

var (
	ErrUserExists   = errors.New("user already exists")
	ErrWeakPassword = errors.New("password is too weak")
)
