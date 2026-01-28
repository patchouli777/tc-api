package user

import (
	"time"
	"twitchy-api/internal/lib/null"
)

type GetRequest struct{}
type GetResponse struct {
	Id              int       `json:"id"`
	Name            string    `json:"name"`
	Pfp             string    `json:"pfp"`
	IsBanned        bool      `json:"is_banned"`
	IsPartner       bool      `json:"is_partner"`
	IsLive          bool      `json:"is_live"`
	FirstLivestream time.Time `json:"first_livestream"`
	LastLivestream  time.Time `json:"last_livestream"`
}

type PostRequest struct {
	Name     string      `json:"name"`
	Password string      `json:"password"`
	Pfp      null.String `json:"pfp"`
}
type PostResponse struct{}

type PatchRequest struct {
	Name      null.String `json:"name"`
	Password  null.String `json:"password"`
	Pfp       null.String `json:"pfp"`
	IsBanned  null.Bool   `json:"is_banned"`
	IsPartner null.Bool   `json:"is_partner"`
}
type PatchResponse struct{}

type DeleteRequest struct {
	Id int `json:"id"`
}
type DeleteResponse struct{}

// type ListRequest struct {
// 	Count           int       `json:"count"`
// 	Page            int       `json:"page"`
// 	Sort            string    `json:"sort"`
// 	Registration    time.Time `json:"registration"`
// 	FirstLivestream time.Time `json:"first_livestream"`
// 	LastLivestream  time.Time `json:"last_livestream"`
// }
// type ListResponse struct {
// 	Users []ListResponseItem
// }
// type ListResponseItem struct {
// 	Id              int
// 	Name            string
// 	IsBanned        bool
// 	IsPartner       bool
// 	FirstLivestream time.Time
// 	LastLivestream  time.Time
// 	Pfp          string
// 	Description     string
// }
