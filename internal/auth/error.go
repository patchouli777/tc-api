package auth

import "errors"

var (
	errWrongCredentials = errors.New("wrong credentials")
	errAlreadyExists    = errors.New("user already exists")
	errNotFound         = errors.New("user not found")
)
