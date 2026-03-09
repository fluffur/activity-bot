package helpers

import "fmt"

func CustomEmoji(id int64, originalEmoji string) string {
	return fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, id, originalEmoji)
}
