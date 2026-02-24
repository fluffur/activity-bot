package call

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"context"
	"math/rand"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const (
	MentionTypeEmoji = 1 << iota
	MentionTypeName
	MentionTypeRole
)

const defaultMentionsPerMessage = 5

var callEmojis = []string{
	"🔔", "📢", "📣", "⚡️", "✨", "🌟", "🔥", "🌈", "☄️", "🚀",
	"💎", "🧿", "🔮", "🍀", "🌸", "🌺", "🌼", "🌻", "🌿", "🍃",
	"🌍", "🌙", "🔆", "🎵", "🎶", "🎨", "🎭", "🎪", "🎬", "🎤",
	"🏆", "🏅", "🎖", "🎟", "🧘", "🧩", "🪁", "🛰", "⚓️", "🛸",
	"💫", "⭐️", "🌠", "🌌", "🪐", "🌊", "💥", "🎇", "🎆", "🕊",
	"👑", "💖", "💙", "💜", "🤍", "💛", "🧡", "❤️‍🔥", "💗", "💞",
	"🪄", "🎀", "🦋", "🐚", "🌷", "🌹", "🌾", "🍓", "🍒", "🍇",
	"🥂", "🍹", "🧁", "🍩", "🍪", "🌞", "🌤", "⛅️", "🌅", "🌄",
	"🌀", "💠", "🧡", "💚", "🤎", "🖤", "🩵", "🩷", "🪻", "🪷",
}

type Service struct {
	repo          chat.Repository
	memberService *member.Service
}

func NewService(repo chat.Repository, memberService *member.Service) *Service {
	return &Service{repo, memberService}
}

func (s *Service) Call(ctx context.Context, b *gotgbot.Bot, tgCtx *ext.Context, message string) error {
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

	chatSettings, err := s.repo.GetChat(ctx, tgCtx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	mentionsLimit := int(chatSettings.MentionsPerMessage)
	if mentionsLimit <= 0 {
		mentionsLimit = defaultMentionsPerMessage
	}

	if message == "" {
		message = chatSettings.WelcomeCallMessage
	}

	for i := 0; i < len(members); i += mentionsLimit {
		end := i + mentionsLimit
		if end > len(members) {
			end = len(members)
		}

		var sb strings.Builder
		if message != "" {
			sb.WriteString(message)

			if chatSettings.MentionTypes != 0 {
				sb.WriteString("\n\n")
			}
		}

		for j, m := range members[i:end] {
			var parts []string
			emptyStr := "​"
			if j == 0 && message == "" {
				emptyStr = "ㅤ"
			}

			emoji := callEmojis[rand.Intn(len(callEmojis))]

			if chatSettings.MentionTypes&MentionTypeEmoji > 0 {
				parts = append(parts, emoji)
			}
			if chatSettings.MentionTypes&MentionTypeName > 0 {
				parts = append(parts, m.User.FirstName)
			}
			if chatSettings.MentionTypes&MentionTypeRole > 0 && m.CustomTitle != "" {
				parts = append(parts, "("+m.CustomTitle+")")
			}

			if len(parts) == 0 {
				parts = append(parts, emptyStr)
			}

			title := strings.Join(parts, " ")
			if strings.TrimSpace(title) == "" {
				title = emptyStr
			}
			sb.WriteString(helpers.Mention(m.User.ID, title))
			if j < len(members[i:end])-1 {
				sb.WriteString(" ")
			}
		}

		if len(tgCtx.EffectiveMessage.Photo) > 0 {
			lastPhoto := tgCtx.EffectiveMessage.Photo[len(tgCtx.EffectiveMessage.Photo)-1]
			if _, err := b.SendPhoto(tgCtx.EffectiveChat.Id, gotgbot.InputFileByID(lastPhoto.FileId), &gotgbot.SendPhotoOpts{
				ParseMode:       gotgbot.ParseModeHTML,
				Caption:         sb.String(),
				HasSpoiler:      tgCtx.EffectiveMessage.HasMediaSpoiler,
				ReplyParameters: replyParams,
			}); err != nil {
				return err
			}
		} else {
			if _, err := tgCtx.EffectiveMessage.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
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

func (s *Service) SetMentionsPerMessage(ctx context.Context, chatID int64, count int32) error {
	return s.repo.SetMentionsPerMessage(ctx, chatID, count)
}

func (s *Service) SetMentionTypes(ctx context.Context, chatID int64, types int32) error {
	return s.repo.SetMentionTypes(ctx, chatID, types)
}
