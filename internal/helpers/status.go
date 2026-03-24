package helpers

import (
	"activity-bot/internal/model"
)

func StatusEmoji(status model.Status) string {
	customEmojis := []string{
		CustomEmoji(5368832132158337100, "0️⃣"),
		CustomEmoji(5366311523226496082, "1️⃣"),
		CustomEmoji(5366536356174510823, "2️⃣"),
		CustomEmoji(5366525906519076469, "3️⃣"),
		CustomEmoji(5368367657215078446, "4️⃣"),
		CustomEmoji(5368552151830244939, "5️⃣"),
	}
	if status < 0 || status > 5 {
		return ""
	}
	return customEmojis[status]
}

func StatusEmojiId(status model.Status) string {
	customEmojis := []string{
		"5368832132158337100",
		"5366311523226496082",
		"5366536356174510823",
		"5366525906519076469",
		"5368367657215078446",
		"5368552151830244939",
	}
	if status < 0 || status > 5 {
		return ""
	}
	return customEmojis[status]
}

func StatusEmojiPlain(status model.Status) string {
	emojis := []string{"0️⃣", "1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣"}
	if status < 0 || status > 5 {
		return ""
	}
	return emojis[status]
}
