package livestream

type LivestreamCategory struct {
	Link string `json:"link"`
	Name string `json:"name"`
}
type GetRequest struct {
	Channel string `json:"channel"`
}
type GetResponse struct {
	Id            int                `json:"id"`
	Username      string             `json:"username"`
	Avatar        string             `json:"avatar"`
	Title         string             `json:"title"`
	StartedAt     int                `json:"started_at"`
	IsLive        bool               `json:"is_live"`
	IsMultistream bool               `json:"is_multistream"`
	Thumbnail     string             `json:"thumbnail"`
	IsFollowing   bool               `json:"is_following"`
	IsSubscriber  bool               `json:"is_subscriber"`
	Viewers       int32              `json:"viewers"`
	Category      LivestreamCategory `json:"category"`
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
	Title        *string `json:"title"`
	CategoryLink *string `json:"category_link"`
}
type PatchResponse struct {
	Status bool
}

type DeleteRequest struct{}
type DeleteResponse struct {
	Status bool `json:"status"`
}

type ListRequest struct {
	Page  int `json:"page"`
	Count int `json:"count"`
}
type ListResponse struct {
	Livestreams []ListResponseItem `json:"livestreams"`
}
type ListResponseItem struct {
	Username      string             `json:"username"`
	Title         string             `json:"title"`
	Avatar        string             `json:"avatar"`
	StartedAt     int                `json:"started_at"`
	IsLive        bool               `json:"is_live"`
	IsMultistream bool               `json:"is_multistream"`
	Thumbnail     string             `json:"thumbnail"`
	Viewers       int32              `json:"viewers"`
	Category      LivestreamCategory `json:"category"`
	// IsPartner     bool   `json:"is_partner"`
}
