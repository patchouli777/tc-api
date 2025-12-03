package channel

import "time"

type Channel struct {
	Name            string
	IsBanned        bool
	IsPartner       bool
	FirstLivestream time.Time
	LastLivestream  time.Time
	Description     string
	Links           []string
	Tags            []string
}

type ChannelUpdate struct {
	Name        *string
	IsBanned    *bool
	IsPartner   *bool
	Description *string
	Links       []string
	Tags        []string
}
