package domain

import (
	"main/internal/lib/null"
	"time"
)

type Channel struct {
	Name            string
	IsBanned        bool
	IsPartner       bool
	FirstLivestream time.Time
	LastLivestream  time.Time
	Description     string
	Background      string
	Links           []ChannelLink
	Tags            []ChannelTag
}

type ChannelLink struct {
	Id   int32
	Name string
	Link string
}

type ChannelTag struct {
	Id   int32
	Name string
}

type ChannelUpdate struct {
	Name null.String
	// IsBanned    null.Bool
	// IsPartner   null.Bool
	Description null.String
	Links       []ChannelLink
	Tags        []ChannelTag
}
