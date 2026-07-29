package genshin

type Category struct {
	Name  string
	Emoji string
	Roles []Role
}

type Role struct {
	Name    string
	Emoji   string
	Aliases []string
}
