package user

import "time"

type GetRequest struct{}
type GetResponse struct {
	Id              int       `json:"id"`
	Name            string    `json:"name"`
	Avatar          string    `json:"avatar"`
	IsBanned        bool      `json:"is_banned"`
	IsPartner       bool      `json:"is_partner"`
	FirstLivestream time.Time `json:"first_livestream"`
	LastLivestream  time.Time `json:"last_livestream"`
}
type GetResponseNotFound struct {
	Result string `json:"result"`
}

type PostRequest struct {
	Name     *string `json:"name"`
	Password *string `json:"password"`
	Avatar   *string `json:"avatar"`
}
type PostResponse struct {
	Success bool `json:"success"`
}

// TODO: test if pointers work
// change everything to pointeers if they are
type PatchRequest struct {
	Name           *string    `json:"name"`
	Password       *string    `json:"password"`
	Avatar         *string    `json:"avatar"`
	IsBanned       *bool      `json:"is_banned"`
	IsPartner      *bool      `json:"is_partner"`
	LastLivestream *time.Time `json:"last_livestream"`
}
type PatchResponse struct {
	Success bool `json:"success"`
}

type DeleteRequest struct {
	UserId int `json:"user_id"`
}
type DeleteResponse struct {
	Success bool `json:"success"`
}

// TODO: курсорная пагинация
type ListRequest struct {
	Count           int       `json:"count"`
	Page            int       `json:"page"`
	Sort            string    `json:"sort"`
	Registration    time.Time `json:"registration"`
	FirstLivestream time.Time `json:"first_livestream"`
	LastLivestream  time.Time `json:"last_livestream"`
}
type ListResponse struct {
	Users []ListResponseItem
}
type ListResponseItem struct {
	Id              int
	Name            string
	IsBanned        bool
	IsPartner       bool
	FirstLivestream time.Time
	LastLivestream  time.Time
	Avatar          string
	Description     string
}
