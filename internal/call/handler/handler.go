package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"fmt"
	"log/slog"
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

func (h *Handler) Call(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	message := cctx.FirstArgument()
	var replyParams *gotgbot.ReplyParameters
	if ctx.EffectiveMessage.ReplyToMessage != nil {
		replyParams = &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.ReplyToMessage.MessageId,
			ChatId:    ctx.EffectiveChat.Id,
		}
	}
	if _, err := h.memberService.SyncChatMembers(ctx.EffectiveChat.Id); err != nil {
		slog.Error("Failed to sync chat members", "error", err)
		return err
	}

	users, err := h.memberService.GetChatMembers(ctx.EffectiveChat.Id)
	if err != nil {
		slog.Error("Failed to get chat members", "error", err)
		return err
	}

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
				ParseMode:       gotgbot.ParseModeHTML,
				Caption:         sb.String(),
				ReplyParameters: replyParams,
			}); err != nil {
				return err
			}
		} else {
			if _, err := b.SendMessage(ctx.EffectiveChat.Id, sb.String(), &gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyParameters: replyParams,
			}); err != nil {
				return err
			}
		}

	}

	return nil
}
