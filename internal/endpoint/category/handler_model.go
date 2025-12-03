package category

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
	Thumbnail string   `json:"thumbnail"`
	IsSafe    bool     `json:"is_safe"`
	Name      string   `json:"name"`
	Link      string   `json:"link"`
	Viewers   int32    `json:"viewers"`
	Tags      []string `json:"tags"`
}

type PostRequest struct {
	Thumbnail string   `json:"thumbnail"`
	Name      string   `json:"name"`
	Link      string   `json:"link"`
	Tags      []string `json:"tags"`
	IsSafe    bool     `json:"is_safe"`
}

// type PostResponse struct {
// 	Success bool
// 	Error   string
// }

type PatchRequest struct {
	Thumbnail *string   `json:"thumbnail"`
	IsSafe    *bool     `json:"is_safe"`
	Name      *string   `json:"name"`
	Link      *string   `json:"link"`
	Tags      *[]string `json:"tags"`
}

// type PatchResponse struct{}

type DeleteRequest struct{}

// type DeleteResponse struct{}
