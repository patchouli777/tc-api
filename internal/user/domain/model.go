package user

import (
	"time"
	"twitchy-api/internal/lib/null"
)

type User struct {
	Id              int32
	Name            string
	Password        string
	IsBanned        bool
	IsPartner       bool
	FirstLivestream time.Time
	LastLivestream  time.Time
	Pfp             string
	Description     string
	Links           []string
	Tags            []string
}

type UserCreate struct {
	Name     string
	Password string
	Pfp      null.String
}

type UserUpdate struct {
	Name      null.String
	Password  null.String
	Pfp       null.String
	IsBanned  null.Bool
	IsPartner null.Bool
}

type UserList struct {
	Count           int
	Page            int
	Sort            string
	Registration    time.Time
	FirstLivestream time.Time
	LastLivestream  time.Time
}
