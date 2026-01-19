package domain

import "errors"

var (
	ErrWrongCredentials = errors.New("wrong credentials")
	ErrAlreadyExists    = errors.New("user already exists")
	ErrNotFound         = errors.New("user not found")
)
