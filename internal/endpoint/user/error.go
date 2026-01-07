package user

import "errors"

var (
	errUserExists       = errors.New("user already exists")
	errWeakPassword     = errors.New("password is too weak")
	errUsernameRequired = errors.New("username required")
	errPasswordRequired = errors.New("password required")
)
