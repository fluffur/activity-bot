package call

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat/member"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
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
	adminService  *admin.Service
	memberService *member.Service
}

func NewHandler(adminService *admin.Service, memberService *member.Service) *Handler {
	return &Handler{adminService, memberService}
}

func (h *Handler) Call(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	var message string
	if len(cctx.Args) != 0 {
		message = cctx.Args[0]
	}

	dbMembers, err := h.memberService.GetChatMembers(ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("GetChatMembers", err)

	}

	admins, err := b.GetChatAdministrators(ctx.EffectiveChat.Id, nil)
	if err != nil {
		log.Println("GetChatAdministrators", err)
		_, err := ctx.EffectiveMessage.Reply(b, "Не удалось созвать пользователей", nil)
		return err
	}

	var tgUsers []gotgbot.User
	for _, a := range admins {
		if a.GetUser().IsBot {
			continue
		}
		tgUsers = append(tgUsers, a.GetUser())
	}

	users := mergeUsers(dbMembers, tgUsers)

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
			sb.WriteString(fmt.Sprintf("%s ", helpers.Mention(user, emoji)))
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

func mergeUsers(dbMembers []model.ChatMember, tgUsers []gotgbot.User) []model.User {
	usersMap := make(map[int64]model.User)

	for _, m := range dbMembers {
		usersMap[m.User.ID] = m.User
	}

	for _, u := range tgUsers {
		if u.IsBot {
			continue
		}

		var username *string
		if u.Username != "" {
			username = &u.Username
		}

		usersMap[u.Id] = model.User{
			ID:        u.Id,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Username:  username,
		}
	}

	result := make([]model.User, 0, len(usersMap))
	for _, u := range usersMap {
		result = append(result, u)
	}

	return result
}
