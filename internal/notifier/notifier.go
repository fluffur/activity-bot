package notifier

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/summon"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

type UsernameChangedNotifier struct {
	translator           *i18n.Translator
	chatMemberRepository chatmember.Repository
}

func NewUsernameChangedNotifier(t *i18n.Translator, cmr chatmember.Repository) *UsernameChangedNotifier {
	return &UsernameChangedNotifier{
		translator:           t,
		chatMemberRepository: cmr,
	}
}

func (n *UsernameChangedNotifier) NotifyUsernameChanged(c *botapi.Context, oldUsername, newUsername string) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("notify username changed: %w", err)
	}

	cm, err := cctx.ChatMember(c.Context)
	if err != nil {
		return fmt.Errorf("notify username changed: %w", err)
	}

	args := i18n.SystemUsernameChangedMaleArgs(
		tghtml.Code("@"+newUsername),
		tghtml.Code("@"+oldUsername),
		tghtml.MemberLink(n.translator, ch, cm),
	)

	text := n.translator.TIf(
		ch.Lang, cm.IsMale(), i18n.System.UsernameChangedMale, i18n.System.UsernameChangedFemale, args, args,
	)

	admins, err := n.chatMemberRepository.ListAdmins(c.Context, ch.ID, chatmember.StatusModerator)
	if err != nil {
		return fmt.Errorf("notify username changed list admins: %w", err)
	}
	if len(admins) > 0 {
		mentions := summon.BuildMentions(admins, ch.MentionTypes)

		if ch.MentionTypes != 0 {
			text += "\n\n"
		}

		text += strings.Join(
			mentions,
			summon.MentionSeparator(ch.MentionTypes),
		)
	}

	_, err = c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}
