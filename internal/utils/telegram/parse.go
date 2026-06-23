package telegram

import (
	"activity-bot/internal/chatmember"
	"strings"

	"github.com/gotd/log"

	"github.com/gotd/botapi"
)

func ParseMentionedMembers(
	c *botapi.Context,
	chatMemberRepo chatmember.Repository,
	chatID int64,
	entities []botapi.MessageEntity,
	text string,
) []chatmember.ChatMember {
	var members []chatmember.ChatMember
	for _, entity := range entities {
		switch entity.Type {
		case botapi.EntityMention:
			username := strings.TrimPrefix(entityText(text, entity.Offset, entity.Length), "@")

			cm, err := chatMemberRepo.GetByUsername(c.Context, chatID, username)
			if err != nil {
				log.For(c.Bot.Logger()).Warn(c.Context, "entity mention get by username", log.Error(err))
				continue
			}
			members = append(members, cm)

		case botapi.EntityTextMention:
			cm, err := chatMemberRepo.Get(c.Context, chatID, entity.User.ID)
			if err != nil {
				log.For(c.Bot.Logger()).Warn(c.Context, "entity text mention get by username", log.Error(err))
				continue
			}

			members = append(members, cm)
		case botapi.EntityURL:
			link := entityText(text, entity.Offset, entity.Length)

			if strings.HasPrefix(link, "t.me/") || strings.HasPrefix(link, "https://t.me/") {
				link = strings.TrimPrefix(strings.TrimPrefix(link, "https://"), "t.me/")
				username := strings.TrimPrefix(link, "@")

				cm, err := chatMemberRepo.GetByUsername(c.Context, chatID, username)
				if err != nil {
					log.For(c.Bot.Logger()).Warn(c.Context, "entity link get by username", log.Error(err))
					continue
				}

				members = append(members, cm)
			}
		}
	}

	return members
}

func entityText(text string, offset, length int) string {
	runes := []rune(text)

	if offset < 0 || offset >= len(runes) {
		return ""
	}

	end := offset + length
	if end > len(runes) {
		end = len(runes)
	}

	return string(runes[offset:end])
}
