package livestream

import "errors"

var (
	ErrAlreadyStarted = errors.New("livestream already started")
	ErrAlreadyEnded   = errors.New("livestream already ended")
	ErrNotFound       = errors.New("livestream is not found")
)
