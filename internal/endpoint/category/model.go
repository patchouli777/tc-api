package category

import (
	"encoding/json"
	"main/internal/lib/null"
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

type CategoryUpdate struct {
	IsSafe    null.Bool
	Thumbnail null.String
	Name      null.String
	Link      null.String
	Tags      null.Array[int32]
}

type CategoryCreate struct {
	Thumbnail string
	Name      string
	Link      string
	Tags      []int32
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
