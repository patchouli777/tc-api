package follow

type FollowingListExtendedItem struct {
	Name     string
	Pfp      string
	Viewers  int
	Title    string
	Category string
	IsLive   bool
}

type FollowerListItem struct {
	Name string
	Pfp  string
}
