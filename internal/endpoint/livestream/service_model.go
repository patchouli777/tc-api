package livestream

import (
	"encoding/json"
)

type Livestream struct {
	Id        int32    `redis:"id"`
	Title     string   `redis:"title"`
	Thumbnail string   `redis:"thumbnail"`
	Viewers   int32    `redis:"viewers"`
	StartedAt int64    `redis:"started_at"`
	User      User     `redis:"user"`
	Category  Category `redis:"category"`
}

type LivestreamUpdate struct {
	Title        *string
	Thumbnail    *string
	Viewers      *int32
	CategoryLink *string
}

type LivestreamCreate struct {
	Title        string
	CategoryLink string
	Username     string
}

type Category struct {
	Name string `redis:"name"`
	Link string `redis:"link"`
}

type CategoryUpdate struct {
	CategoryLink string
	CategoryName string
}

func (m Category) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Category) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

type User struct {
	Name   string `redis:"name"`
	Avatar string `redis:"avatar"`
}

func (m User) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}
