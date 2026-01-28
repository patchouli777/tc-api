package domain

import (
	"encoding/json"
	"strconv"
	"twitchy-api/internal/lib/null"
	api "twitchy-api/pkg/api/livestream"
)

type Livestream struct {
	Id           int    `redis:"id"`
	Title        string `redis:"title"`
	Thumbnail    string `redis:"thumbnail"`
	Viewers      int    `redis:"viewers"`
	StartedAt    int    `redis:"started_at"`
	UserId       int    `redis:"user:id"`
	UserName     string `redis:"user:name"`
	UserPfp      string `redis:"user:pfp"`
	CategoryId   int    `redis:"category:id"`
	CategoryName string `redis:"category:name"`
	CategoryLink string `redis:"category:link"`
}

func (l *Livestream) ToGetResponse() api.GetResponse {
	return api.GetResponse{
		Id:        l.Id,
		StartedAt: l.StartedAt,
		Viewers:   l.Viewers,
		Channel: api.LivestreamChannel{
			Id:         strconv.Itoa(l.UserId),
			Username:   l.UserName,
			ProfilePic: l.UserPfp,
		},
		Category: api.LivestreamCategory{
			Id:   l.CategoryId,
			Link: l.CategoryLink,
			Name: l.CategoryName,
		},
		Title:         l.Title,
		IsLive:        true,
		IsMultistream: false,
		Thumbnail:     l.Thumbnail,
		IsFollowing:   false,
		IsSubscriber:  false,
	}
}

func (l *Livestream) ToListResponseItem() api.ListResponseItem {
	return api.ListResponseItem{
		Id: l.Id,
		Channel: api.LivestreamChannel{
			Id:         strconv.Itoa(l.UserId),
			Username:   l.UserName,
			ProfilePic: l.UserPfp,
		},
		Category: api.LivestreamCategory{
			Id:   l.CategoryId,
			Name: l.CategoryName,
			Link: l.CategoryLink,
		},
		StartedAt: l.StartedAt,
		Thumbnail: l.Thumbnail,
		Viewers:   l.Viewers,
		Title:     l.Title,
	}
}

type LivestreamUpdate struct {
	Title      null.String
	CategoryId null.Int
}

type LivestreamCreate struct {
	Username string
}

type LivestreamSearch struct {
	CategoryId string
	Category   string
	Page       int
	Count      int
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
	Name string `redis:"name"`
	Pfp  string `redis:"pfp"`
}

func (m User) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}
