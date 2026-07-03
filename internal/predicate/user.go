package predicate

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"context"
	"strings"
	"unicode/utf16"

	"github.com/gotd/botapi"
)

func resolveUserEntities(
	ctx context.Context,
	repo chatmember.Repository,
	chatID int64,
	text string,
	entities []botapi.MessageEntity,
	parsed *cctx.ParsedArgs,
	used []Offset,
) []Offset {
	for _, entity := range entities {
		entityStr := entityTextUTF16(text, entity.Offset, entity.Length)
		if entityStr == "" {
			continue
		}

		u16 := utf16.Encode([]rune(text))
		byteStart := len(string(utf16.Decode(u16[:entity.Offset])))
		byteEnd := byteStart + len(string(utf16.Decode(u16[entity.Offset:entity.Offset+entity.Length])))

		switch entity.Type {
		case botapi.EntityMention:
			username := strings.TrimPrefix(entityStr, "@")

			cm, err := repo.GetByUsername(ctx, chatID, username)
			if err == nil {
				parsed.Users = append(parsed.Users, cm)
				used = append(used, Offset{byteStart, byteEnd})
			}
		case botapi.EntityTextMention:
			if entity.User != nil {
				if cm, err := repo.Get(ctx, chatID, entity.User.ID); err == nil {
					parsed.Users = append(parsed.Users, cm)
					used = append(used, Offset{byteStart, byteEnd})
				}
			}
		case botapi.EntityURL:
			if strings.Contains(entityStr, "t.me/") {
				parts := strings.Split(entityStr, "/")
				username := strings.TrimPrefix(parts[len(parts)-1], "@")

				if cm, err := repo.GetByUsername(ctx, chatID, username); err == nil {
					parsed.Users = append(parsed.Users, cm)
					used = append(used, Offset{byteStart, byteEnd})
				}
			}
		default:
		}
	}

	return used
}
