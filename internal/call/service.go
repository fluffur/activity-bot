package call

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"context"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

const (
	MentionTypeNWSP  = 0
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

var tgMentionRegex = regexp.MustCompile(`(?i)<a\s+href="tg://user\?id=\d+">([^<]+)</a>`)

var mentionRegex = regexp.MustCompile(`(?i)(^|[^A-Za-z0-9_])@([a-zA-Z0-9_]{5,32})`)

func replaceMentionsWithLinks(input string) string {
	input = tgMentionRegex.ReplaceAllString(input, "<a href=\"tg://openmessage?user_id=$2\">$1</a>")
	var sb strings.Builder
	inTag := false
	var textBuf strings.Builder

	flushText := func() {
		if textBuf.Len() > 0 {
			res := mentionRegex.ReplaceAllString(textBuf.String(), "$1<a href=\"https://t.me/$2\">@$2</a>")
			log.Println(textBuf.String(), res)
			sb.WriteString(res)
			textBuf.Reset()
		}
	}

	for _, r := range input {
		if r == '<' {
			flushText()
			inTag = true
			sb.WriteRune(r)
		} else if r == '>' && inTag {
			inTag = false
			sb.WriteRune(r)
		} else if inTag {
			sb.WriteRune(r)
		} else {
			textBuf.WriteRune(r)
		}
	}
	flushText()
	return sb.String()
}

func (s *Service) Call(ctx *cmd.Context, b *gotgbot.Bot, message string) error {
	log.Println(message)
	members, err := s.memberService.GetChatMembers(ctx.StdContext(), ctx.TargetChatID())
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

	chatSettings, err := s.repo.GetChat(ctx.StdContext(), ctx.TargetChatID())
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

	if message != "" {
		message = replaceMentionsWithLinks(message)
	}
	log.Println(message)

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

		if len(ctx.EffectiveMessage.Photo) > 0 {
			lastPhoto := ctx.EffectiveMessage.Photo[len(ctx.EffectiveMessage.Photo)-1]
			if _, err := b.SendPhoto(ctx.TargetChatID(), gotgbot.InputFileByID(lastPhoto.FileId), &gotgbot.SendPhotoOpts{
				ParseMode:       gotgbot.ParseModeHTML,
				Caption:         sb.String(),
				HasSpoiler:      ctx.EffectiveMessage.HasMediaSpoiler,
				ReplyParameters: replyParams,
			}); err != nil {
				return err
			}
		} else {
			if _, err := ctx.EffectiveMessage.Reply(b, sb.String(), &gotgbot.SendMessageOpts{
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
