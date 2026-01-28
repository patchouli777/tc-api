package livestream

import "twitchy-api/internal/lib/null"

type LivestreamCategory struct {
	Id   int    `json:"id"`
	Link string `json:"link"`
	Name string `json:"name"`
}

type GetRequest struct {
	Channel string `json:"channel"`
}
type GetResponse struct {
	Id            int                `json:"id"`
	Title         string             `json:"title"`
	StartedAt     int                `json:"started_at"`
	IsLive        bool               `json:"is_live"`
	IsMultistream bool               `json:"is_multistream"`
	Thumbnail     string             `json:"thumbnail"`
	IsFollowing   bool               `json:"is_following"`
	IsSubscriber  bool               `json:"is_subscriber"`
	Viewers       int                `json:"viewers"`
	Channel       LivestreamChannel  `json:"channel"`
	Category      LivestreamCategory `json:"category"`
}

type LivestreamChannel struct {
	Id         string `json:"id"`
	Username   string `json:"username"`
	ProfilePic string `json:"profile_pic"`
}
type ListRequest struct {
	Page  int `json:"page"`
	Count int `json:"count"`
}
type ListResponse struct {
	Livestreams []ListResponseItem `json:"livestreams"`
}
type ListResponseItem struct {
	Id        int                `json:"id"`
	Title     string             `json:"title"`
	StartedAt int                `json:"started_at"`
	Thumbnail string             `json:"thumbnail"`
	Viewers   int                `json:"viewers"`
	Channel   LivestreamChannel  `json:"channel"`
	Category  LivestreamCategory `json:"category"`
	// IsMultistream bool               `json:"is_multistream"`
	// IsPartner     bool   `json:"is_partner"`
}

type PostRequest struct {
	Title        string `json:"title"`
	CategoryLink string `json:"category_link"`
}
type PostResponse struct {
	Username  string `json:"username"`
	Category  string `json:"category"`
	Title     string `json:"title"`
	StartedAt int64  `json:"started_at"`
	Viewers   int32  `json:"viewers"`
}

type PatchRequest struct {
	// Channel      null.String `json:"channel"`
	Title      null.String `json:"title"`
	CategoryId null.Int    `json:"category_id"`
}
type PatchResponse struct {
	Status bool
}

type DeleteRequest struct{}
type DeleteResponse struct {
	Status bool `json:"status"`
}
