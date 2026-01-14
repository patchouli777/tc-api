package channel

import "time"

type Tag struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
}

type Link struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}
type GetRequest struct{}
type GetResponse struct {
	Id              int       `json:"id"`
	Name            string    `json:"name"`
	IsBanned        bool      `json:"is_banned"`
	IsPartner       bool      `json:"is_partner"`
	FirstLivestream time.Time `json:"first_livestream"`
	LastLivestream  time.Time `json:"last_livestream"`
	Description     string    `json:"description"`
	Background      string    `json:"background"`
	Links           []Link    `json:"links"`
	Tags            []Tag     `json:"tags"`
}

type ListRequest struct{}
type ListResponse struct{}
type ListResponseItem struct{}

type PostRequest struct{}
type PostResponse struct{}

type PatchRequest struct {
}
type PatchResponse struct{}

type DeleteRequest struct{}
type DeleteResponse struct{}
