package call

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const mentionsPerMessage = 5

var emojis = []string{
	"🔔", "👋", "⚡", "📢", "🔥",
	"🟡", "🟢", "🔹", "🔸", "⭐",
	"✨", "🏆", "🎯", "🚨", "🛎️",
	"🥁", "📣", "🎉", "🟠", "🟣",
}

type Handler struct {
	adminService *admin.Service
}

func NewHandler(adminService *admin.Service) *Handler {
	return &Handler{adminService}
}

func (h *Handler) Call(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	var message string
	if len(cctx.Args) != 0 {
		message = cctx.Args[0]
	}

	admins, err := b.GetChatAdministrators(ctx.EffectiveChat.Id, nil)
	if err != nil {
		log.Println("GetChatAdministrators", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось созвать пользователей", nil)
		return err
	}

	var users []gotgbot.User
	for _, a := range admins {
		if a.GetUser().IsBot {
			continue
		}
		users = append(users, a.GetUser())
	}

	for i := 0; i < len(users); i += mentionsPerMessage {
		end := i + mentionsPerMessage
		if end > len(users) {
			end = len(users)
		}

		var sb strings.Builder

		if i == 0 && message != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", message))
		}

		for _, user := range users[i:end] {
			emoji := emojis[rand.Intn(len(emojis))]
			sb.WriteString(fmt.Sprintf("%s ", helpers.Mention(helpers.MapUser(&user), emoji)))
		}

		photos := ctx.EffectiveMessage.Photo
		if len(photos) != 0 {
			lastPhoto := photos[len(photos)-1]
			if _, err := b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByID(lastPhoto.FileId), &gotgbot.SendPhotoOpts{
				ParseMode: gotgbot.ParseModeHTML,
				Caption:   sb.String(),
			}); err != nil {
				return err
			}
		} else {
			if _, err := b.SendMessage(ctx.EffectiveChat.Id, sb.String(), &gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
			}); err != nil {
				return err
			}
		}

	}

	return nil
}
