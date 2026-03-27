package helpers

import (
	"activity-bot/internal/model"
)

func StatusEmoji(status model.Status) string {
	return CustomEmoji(StatusEmojiID(status), StatusEmojiPlain(status))
}

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
