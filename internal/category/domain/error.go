package domain

import "errors"

var (
	ErrAlreadyExists = errors.New("category already exists")
	ErrNotFound      = errors.New("category not found")
	ErrEmptyNameLink = errors.New("category link and/or name is empty")
)
