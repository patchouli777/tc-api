package domain

import (
	"encoding/json"
	"main/internal/lib/null"
	api "main/pkg/api/category"
)

type CategoryTag struct {
	Id   int32  `json:"id" redis:"id"`
	Name string `json:"name" redis:"name"`
}

// kekw? https://github.com/redis/go-redis/issues/1554
type CategoryTags []CategoryTag

type Category struct {
	Id        int32        `redis:"id"`
	IsSafe    bool         `redis:"is_safe"`
	Thumbnail string       `redis:"thumbnail"`
	Name      string       `redis:"name"`
	Link      string       `redis:"link"`
	Viewers   int32        `redis:"viewers"`
	Tags      CategoryTags `redis:"tags"`
}

func (c *Category) ToListResponseItem() api.ListResponseItem {
	tags := make([]api.CategoryTag, len(c.Tags))
	for i, t := range c.Tags {
		tags[i] = api.CategoryTag{Id: int(t.Id), Name: t.Name}
	}

	return api.ListResponseItem{
		Name:      c.Name,
		Thumbnail: c.Thumbnail,
		Link:      c.Link,
		Viewers:   int(c.Viewers),
		Tags:      tags,
		IsSafe:    c.IsSafe,
	}
}

func (c *Category) ToGetResponse() api.GetResponse {
	tags := make([]api.CategoryTag, len(c.Tags))
	for i, t := range c.Tags {
		tags[i] = api.CategoryTag{Id: int(t.Id), Name: t.Name}
	}

	return api.GetResponse{
		Category: api.Category{
			Id:        int(c.Id),
			IsSafe:    c.IsSafe,
			Thumbnail: c.Thumbnail,
			Name:      c.Name,
			Link:      c.Link,
			Viewers:   int(c.Viewers),
			Tags:      tags,
		}}
}

type CategoryUpdate struct {
	IsSafe    null.Bool
	Thumbnail null.String
	Name      null.String
	Link      null.String
	Tags      null.Array[int]
}

type CategoryCreate struct {
	Thumbnail string
	Name      string
	Link      string
	Tags      []int
}

type CategoryFilter struct {
	Page  uint32
	Count uint64
	Sort  string
}

func (ct CategoryTag) MarshalBinary() ([]byte, error) {
	return json.Marshal(ct)
}

func (cta CategoryTags) MarshalBinary() ([]byte, error) {
	return json.Marshal(cta)
}
