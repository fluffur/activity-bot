package helpers

import (
	"activity-bot/internal/model"
	"fmt"
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

func WriteMention(eb *entity.Builder, id int64, value string) {
	if value == "" {
		value = "​"
	}
	eb.MentionName(value, &tg.InputUser{UserID: id})
}

func WriteMemberMention(eb *entity.Builder, member model.ChatMember) {
	title := member.Tag
	if title == "" {
		title = member.User.FirstName
	}
	eb.MentionName(title, &tg.InputUser{UserID: member.User.ID})
}

func WriteUserMention(eb *entity.Builder, u model.User) {
	eb.MentionName(u.FirstName, &tg.InputUser{UserID: u.ID})
}

func WriteRoleEmojiLink(eb *entity.Builder, cm model.ChatMember, emojisEnabled bool) {
	if emojis := MemberDisplayEmojis(cm, emojisEnabled); len(emojis) != 0 {
		DisplayEmoji(eb, emojis)
		eb.Plain(" ")
	}
	eb.TextURL(MemberDisplayName(cm), UserLink(cm.User))
}

func WriteRoleEmojiMention(eb *entity.Builder, cm model.ChatMember, emojisEnabled bool) {
	if emojis := MemberDisplayEmojis(cm, emojisEnabled); len(emojis) != 0 {
		DisplayEmoji(eb, emojis)
		eb.Plain(" ")
	}
	WriteMemberMention(eb, cm)
}
