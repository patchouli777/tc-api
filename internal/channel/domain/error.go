package domain

import "errors"

var (
	ErrNotPresent = errors.New("channel is not present in the request")
)
