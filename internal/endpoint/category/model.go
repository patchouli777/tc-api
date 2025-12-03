package category

type Category struct {
	Id        int32    `redis:"id"`
	IsSafe    bool     `redis:"is_safe"`
	Thumbnail string   `redis:"thumbnail"`
	Name      string   `redis:"mame"`
	Link      string   `redis:"link"`
	Viewers   int32    `redis:"viewers"`
	Tags      []string `redis:"-" json:"-"`
}

type CategoryUpdate struct {
	IsSafe    *bool
	Thumbnail *string
	Name      *string
	Link      *string
	Viewers   *int32
	Tags      *[]string
}

type CategoryCreate struct {
	Thumbnail string
	Name      string
	Link      string
	Viewers   int
	Tags      []string
}

type CategoryFilter struct {
	Page  uint32
	Count uint64
	Sort  string
}
