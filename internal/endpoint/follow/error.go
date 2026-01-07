package follow

import "errors"

var (
	errNoFollowed = errors.New("followed username is not present")
	errNoFollower = errors.New("follower username is not present")
)
