package category

import (
	"main/internal/lib/null"
)

type CategoryTag struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
}
type Category struct {
	Id        int32         `json:"id"`
	IsSafe    bool          `json:"is_safe"`
	Thumbnail string        `json:"thumbnail"`
	Name      string        `json:"name"`
	Link      string        `json:"link"`
	Viewers   int32         `json:"viewers"`
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
	Thumbnail string        `json:"thumbnail"`
	IsSafe    bool          `json:"is_safe"`
	Name      string        `json:"name"`
	Link      string        `json:"link"`
	Viewers   int32         `json:"viewers"`
	Tags      []CategoryTag `json:"tags"`
}

type PostRequest struct {
	Thumbnail string  `json:"thumbnail"`
	Name      string  `json:"name"`
	Link      string  `json:"link"`
	Tags      []int32 `json:"tags"`
	IsSafe    bool    `json:"is_safe"`
}
type PostResponse struct{}

type PatchRequest struct {
	Thumbnail null.String       `json:"thumbnail"`
	IsSafe    null.Bool         `json:"is_safe"`
	Name      null.String       `json:"name"`
	Link      null.String       `json:"link"`
	Tags      null.Array[int32] `json:"tags"`
}
type PatchResponse struct{}

type DeleteRequest struct{}
type DeleteResponse struct{}
