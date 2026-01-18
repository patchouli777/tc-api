package channel

import "errors"

var (
	errNotPresent = errors.New("channel is not present in the request")
)
