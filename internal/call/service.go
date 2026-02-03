package call

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"context"
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const mentionsPerMessage = 7

type Service struct {
	repo          chat.Repository
	memberService *member.Service
}

func NewService(repo chat.Repository, memberService *member.Service) *Service {
	return &Service{repo, memberService}
}

func (s *Service) Call(b *gotgbot.Bot, ctx *ext.Context, message string) error {
	if _, err := s.memberService.SyncChatMembers(ctx.EffectiveChat.Id); err != nil {
		return err
	}

	users, err := s.memberService.GetChatMembers(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	var replyParams *gotgbot.ReplyParameters
	if ctx.EffectiveMessage.ReplyToMessage != nil {
		replyParams = &gotgbot.ReplyParameters{
			ChatId:    ctx.EffectiveChat.Id,
			MessageId: ctx.EffectiveMessage.ReplyToMessage.MessageId,
		}
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

		if len(ctx.EffectiveMessage.Photo) > 0 {
			lastPhoto := ctx.EffectiveMessage.Photo[len(ctx.EffectiveMessage.Photo)-1]
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

func (s *Service) SetWelcomeCallMessage(chatID int64, message string) error {
	ctx := context.Background()

	return s.repo.SetWelcomeCallMessage(ctx, chatID, message)

}

func (s *Service) EnableCallOnJoin(chatID int64) error {
	ctx := context.Background()

	return s.repo.UpdateCallOnJoin(ctx, chatID, true)
}

func (s *Service) DisableCallOnJoin(chatID int64) error {
	ctx := context.Background()

	return s.repo.UpdateCallOnJoin(ctx, chatID, false)

}
