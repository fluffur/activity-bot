package roles

type Fandom struct {
	ID         int64
	Name       string
	Categories []Category
}

type Category struct {
	ID    int64
	Name  string
	Roles []Role
}

type Role struct {
	ID      int64
	Name    string
	Emoji   string
	Aliases []string
}
