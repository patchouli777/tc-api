package user

import "errors"

var (
	errNotFound         = errors.New("user not found")
	errAlreadyExists    = errors.New("user already exists")
	errWeakPassword     = errors.New("password is too weak")
	errUsernameRequired = errors.New("username required")
	errPasswordRequired = errors.New("password required")
)
