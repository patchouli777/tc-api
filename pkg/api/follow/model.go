package follow

type GetRequest struct{}
type GetResponse struct {
	IsFollower bool `json:"is_follower"`
}

type ListRequest struct{}
type ListResponse struct {
	FollowList []ListResponseItem `json:"follow_list"`
}
type ListResponseItem struct {
	Name string `json:"name"`
	Pfp  string `json:"pfp"`
}

type ListExtendedResponse struct {
	FollowList []ListExtendedResponseItem `json:"follow_list"`
}
type ListExtendedResponseItem struct {
	Name     string `json:"name"`
	Pfp      string `json:"pfp"`
	Viewers  int    `json:"viewers"`
	Title    string `json:"title"`
	Category string `json:"category"`
	IsLive   bool   `json:"is_live"`
}

type PostRequest struct {
	Follow string `json:"follow"`
}
type PostResponse struct {
	Success bool `json:"success"`
}

type DeleteRequest struct {
	Unfollow string `json:"unfollow"`
}
type DeleteResponse struct {
	Success bool `json:"success"`
}
