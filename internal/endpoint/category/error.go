package category

import "errors"

var (
	errAlreadyExists = errors.New("category already exists")
	errNotFound      = errors.New("category not found")
	errEmptyNameLink = errors.New("category link and/or name is empty")
)
