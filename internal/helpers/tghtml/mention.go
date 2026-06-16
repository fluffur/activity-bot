package tghtml

import "fmt"

func Mention(userID int64, text string) string {
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, userID, text)
}
