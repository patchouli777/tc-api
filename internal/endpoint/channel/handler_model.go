package channel

import "time"

type GetRequest struct{}
type GetResponse struct {
	IsBanned        bool      `json:"is_banned"`
	IsPartner       bool      `json:"is_partner"`
	FirstLivestream time.Time `json:"first_livestream"`
	LastLivestream  time.Time `json:"last_livestream"`
	Description     string    `json:"description"`
	Background      string    `json:"background"`
	Links           []string  `json:"links"`
	Tags            []string  `json:"tags"`
}

type ListRequest struct{}
type ListResponse struct{}
type ListResponseItem struct{}

type PostRequest struct{}
type PostResponse struct{}

type PatchRequest struct{}
type PatchResponse struct{}

type DeleteRequest struct{}
type DeleteResponse struct{}
