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

func (s *Service) Call(ctx context.Context, b *gotgbot.Bot, tgCtx *ext.Context, message string) error {
	if _, err := s.memberService.SyncChatMembers(ctx, tgCtx.EffectiveChat.Id); err != nil {
		return err
	}

	members, err := s.memberService.GetChatMembers(ctx, tgCtx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	var replyParams *gotgbot.ReplyParameters
	if tgCtx.EffectiveMessage.ReplyToMessage != nil {
		replyParams = &gotgbot.ReplyParameters{
			ChatId:    tgCtx.EffectiveChat.Id,
			MessageId: tgCtx.EffectiveMessage.ReplyToMessage.MessageId,
		}
	}

	for i := 0; i < len(members); i += mentionsPerMessage {
		end := i + mentionsPerMessage
		if end > len(members) {
			end = len(members)
		}

		var sb strings.Builder
		if message != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", message))
		}

		for j, m := range members[i:end] {
			if m.User.ID == 1106062335 {
				continue
			}
			sb.WriteString(helpers.Mention(m.User.ID, m.CustomTitle))
			if j < len(members[i:end])-1 {
				sb.WriteString(", ")
			}
		}

		if len(tgCtx.EffectiveMessage.Photo) > 0 {
			lastPhoto := tgCtx.EffectiveMessage.Photo[len(tgCtx.EffectiveMessage.Photo)-1]
			if _, err := b.SendPhoto(tgCtx.EffectiveChat.Id, gotgbot.InputFileByID(lastPhoto.FileId), &gotgbot.SendPhotoOpts{
				ParseMode:       gotgbot.ParseModeHTML,
				Caption:         sb.String(),
				ReplyParameters: replyParams,
			}); err != nil {
				return err
			}
		} else {
			if _, err := b.SendMessage(tgCtx.EffectiveChat.Id, sb.String(), &gotgbot.SendMessageOpts{
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

func (s *Service) SetWelcomeCallMessage(ctx context.Context, chatID int64, message string) error {
	return s.repo.SetWelcomeCallMessage(ctx, chatID, message)
}

func (s *Service) EnableCallOnJoin(ctx context.Context, chatID int64) error {
	return s.repo.UpdateCallOnJoin(ctx, chatID, true)
}

func (s *Service) DisableCallOnJoin(ctx context.Context, chatID int64) error {
	return s.repo.UpdateCallOnJoin(ctx, chatID, false)
}
