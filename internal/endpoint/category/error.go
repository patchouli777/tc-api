package category

import "errors"

var (
	errAlreadyExists = errors.New("category already exists")
	errNotFound      = errors.New("category is not found")
)
