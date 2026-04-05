package helpers

import (
	"activity-bot/internal/model"
	"fmt"
	"html"
	"strings"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

func UserLink(u model.User) string {
	if u.Username != "" {
		return "https://t.me/" + u.Username
	}

	return fmt.Sprintf("tg://openmessage?user_id=%d", u.ID)
}

func Link(username, content string) string {
	return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, username, html.EscapeString(content))
}

func AnyLink(href, content string) string {
	return "<a href=\"" + href + "\">" + content + "</a>"
}

func LinkWithContent(u model.User, content string) string {
	if u.Username != "" {
		return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, u.Username, html.EscapeString(content))
	}

	return fmt.Sprintf(`<a href="tg://openmessage?user_id=%d">%s</a>`, u.ID, html.EscapeString(content))
}

func RoleEmojiLink(cm model.ChatMember) string {
	var emoji string
	if cm.Emoji != "" {
		emoji = cm.Emoji + " "
	} else if cm.User.Emoji != "" {
		emoji = cm.User.Emoji + " "
	}
	return emoji + LinkWithContent(cm.User, MemberDisplayName(cm))

}

func RoleLink(cm model.ChatMember) string {
	return LinkWithContent(cm.User, MemberDisplayName(cm))
}

func MemberDisplayName(cm model.ChatMember) string {
	var displayName string
	if cm.Tag != "" {
		displayName = cm.Tag
	} else {
		fullName := strings.TrimSpace(cm.User.FirstName + " " + cm.User.LastName)
		if fullName == "" {
			fullName = "—"
		}
		displayName = fullName
	}
	return displayName
}

func RoleMentionEmoji(cm model.ChatMember) string {
	var emoji string
	if cm.Emoji != "" {
		emoji = cm.Emoji + " "
	} else if cm.User.Emoji != "" {
		emoji = cm.User.Emoji + " "
	}
	return emoji + Mention(cm.User.ID, MemberDisplayName(cm))

}

func Mention(id int64, value string) string {
	if value == "" {
		value = "—"
	}
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, id, html.EscapeString(value))
}

func TelegramChannelLink(username string) string {
	return fmt.Sprintf("https://t.me/%s", username)
}

func WriteMention(eb *entity.Builder, id int64, value string) {
	if value == "" {
		value = "—"
	}
	eb.MentionName(value, &tg.InputUser{UserID: id})
}

func WriteRoleEmojiLink(eb *entity.Builder, cm model.ChatMember) {
	if cm.Emoji != "" {
		WriteEmoji(eb, cm.Emoji, cm.Emojis)
		eb.Plain(" ")
	} else if cm.User.Emoji != "" {
		WriteEmoji(eb, cm.User.Emoji, cm.User.Emojis)
		eb.Plain(" ")
	}
	eb.TextURL(MemberDisplayName(cm), UserLink(cm.User))
}
