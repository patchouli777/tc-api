package user

import "time"

type User struct {
	Id              int
	Name            string
	Password        string
	IsBanned        bool
	IsPartner       bool
	FirstLivestream time.Time
	LastLivestream  time.Time
	Avatar          string
	Description     string
	Links           []string
	Tags            []string
}

type UserCreate struct {
	Name     string
	Password string
	Avatar   string
}

type UserUpdate struct {
	Name           *string
	Password       *string
	Avatar         *string
	IsBanned       *bool
	IsPartner      *bool
	LastLivestream *time.Time
}

type UserList struct {
	Count           int
	Page            int
	Sort            string
	Registration    time.Time
	FirstLivestream time.Time
	LastLivestream  time.Time
}
