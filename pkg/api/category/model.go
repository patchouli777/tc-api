package category

import (
	"twitchy-api/internal/lib/null"
)

type CategoryTag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type Category struct {
	Id        int           `json:"id"`
	IsSafe    bool          `json:"is_safe"`
	Thumbnail string        `json:"thumbnail"`
	Name      string        `json:"name"`
	Link      string        `json:"link"`
	Viewers   int           `json:"viewers"`
	Tags      []CategoryTag `json:"tags"`
}
type GetRequest struct {
	CategoryLink string `json:"link"`
}
type GetResponse struct {
	Category Category `json:"category"`
}

type ListRequest struct{}
type ListResponse struct {
	Categories []ListResponseItem `json:"categories"`
}
type ListResponseItem struct {
	Id        int           `json:"id"`
	Thumbnail string        `json:"thumbnail"`
	IsSafe    bool          `json:"is_safe"`
	Name      string        `json:"name"`
	Link      string        `json:"link"`
	Viewers   int           `json:"viewers"`
	Tags      []CategoryTag `json:"tags"`
}

type PostRequest struct {
	Thumbnail string `json:"thumbnail"`
	Name      string `json:"name"`
	Link      string `json:"link"`
	Tags      []int  `json:"tags"`
	IsSafe    bool   `json:"is_safe"`
}
type PostResponse struct{}

type PatchRequest struct {
	Thumbnail null.String     `json:"thumbnail"`
	IsSafe    null.Bool       `json:"is_safe"`
	Name      null.String     `json:"name"`
	Link      null.String     `json:"link"`
	Tags      null.Array[int] `json:"tags"`
}
type PatchResponse struct{}

type DeleteRequest struct{}
type DeleteResponse struct{}
