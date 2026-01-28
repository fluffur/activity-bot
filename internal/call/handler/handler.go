package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"fmt"
	"log"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const mentionsPerMessage = 7

type Handler struct {
	memberService *member.Service
}

func New(memberService *member.Service) *Handler {
	return &Handler{memberService}
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

	var tgUsers []gotgbot.MergedChatMember
	for _, a := range admins {
		if a.GetUser().IsBot {
			continue
		}
		tgUsers = append(tgUsers, a.MergeChatMember())
	}

	users := mergeUsers(dbMembers, tgUsers)
	for i := 0; i < len(users); i += mentionsPerMessage {
		end := i + mentionsPerMessage
		if end > len(users) {
			end = len(users)
		}

		var sb strings.Builder

		if message != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", message))
		}

		for j, user := range users[i:end] {
			sb.WriteString(helpers.Mention(user.User.ID, user.CustomTitle))
			if j < len(users[i:end])-1 {
				sb.WriteString(", ")
			}
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

func mergeUsers(dbMembers []model.ChatMember, tgUsers []gotgbot.MergedChatMember) []model.ChatMember {
	usersMap := make(map[int64]model.ChatMember)

	for _, m := range tgUsers {
		if m.User.IsBot {
			continue
		}

		var username *string
		if m.User.Username != "" {
			username = &m.User.Username
		}
		var role string
		if m.GetStatus() == "creator" {
			role = "creator"
		} else {
			role = "member"
		}
		usersMap[m.User.Id] = model.ChatMember{
			User: model.User{
				ID:        m.User.Id,
				FirstName: m.User.FirstName,
				LastName:  m.User.LastName,
				Username:  username,
			},
			CustomTitle: m.CustomTitle,
			Role:        role,
		}
	}

	for _, m := range dbMembers {
		usersMap[m.User.ID] = model.ChatMember{
			User:        m.User,
			ChatID:      m.ChatID,
			ExemptUntil: m.ExemptUntil,
			CustomTitle: m.CustomTitle,
			Role:        m.Role,
		}
	}

	result := make([]model.ChatMember, 0, len(usersMap))
	for _, u := range usersMap {
		result = append(result, u)
	}
	return result
}
