package auth

import "errors"

var (
	errTokenNotFound     = errors.New("token not found")
	errWrongCredentials  = errors.New("wrong credentials")
	errUserAlreadyExists = errors.New("user already exists")
)
