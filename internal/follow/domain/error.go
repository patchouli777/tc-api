package follow

import "errors"

var (
	ErrNoFollowed = errors.New("followed username is not present")
	ErrNoFollower = errors.New("follower username is not present")
)
