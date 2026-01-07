package livestream

import "errors"

var (
	errAlreadyStarted = errors.New("livestream already started")
	errAlreadyEnded   = errors.New("livestream already ended")
	errNotFound       = errors.New("livestream is not found")
	errNoCategory     = errors.New("neither category nor category id is present")
)
