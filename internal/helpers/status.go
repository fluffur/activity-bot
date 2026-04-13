package helpers

import (
	"activity-bot/internal/model"
	"strconv"

	"github.com/gotd/td/telegram/message/entity"
)

func StatusEmojiID(status model.Status) string {
	switch status {
	case model.StatusMember:
		return "5888465618117598264"
	case model.StatusModerator:
		return "5888965496476277736"
	case model.StatusAdmin:
		return "5888565441747492339"
	case model.StatusSeniorAdmin:
		return "5888782384840578820"
	case model.StatusCoOwner:
		return "5888985064347277672"
	case model.StatusOwner:
		return "5888510315842247017"
	}
	return ""
}

func StatusEmojiPlain(status model.Status) string {
	switch status {
	case model.StatusMember:
		return "0️⃣"
	case model.StatusModerator:
		return "1️⃣"
	case model.StatusAdmin:
		return "2️⃣"
	case model.StatusSeniorAdmin:
		return "3️⃣"
	case model.StatusCoOwner:
		return "4️⃣"
	case model.StatusOwner:
		return "5️⃣"
	}
	return ""
}

func WriteStatusEmoji(eb *entity.Builder, status model.Status) {
	id, _ := strconv.ParseInt(StatusEmojiID(status), 10, 64)
	eb.CustomEmoji(StatusEmojiPlain(status), id)
}
