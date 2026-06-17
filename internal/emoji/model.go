package emoji

type EmojiType string

const (
	EmojiTypeUnicode EmojiType = "unicode"
	EmojiTypeCustom  EmojiType = "custom"
)

type Emoji struct {
	ID   int64     `json:"id,omitempty"`
	Type EmojiType `json:"type"`
	Char string    `json:"char"`
}

type Emojis []Emoji
