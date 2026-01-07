package db

import "errors"

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = errors.New("not found")
)
