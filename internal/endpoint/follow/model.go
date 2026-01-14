package follow

type FollowingListExtendedItem struct {
	Name     string
	Avatar   string
	Viewers  int32
	Title    string
	Category string
	IsLive   bool
}

type FollowerListItem struct {
	Name   string
	Avatar string
}
