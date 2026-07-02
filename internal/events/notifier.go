package events

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/summon"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

type UsernameChangedNotifier struct {
	chatMemberRepository chatmember.Repository
}

func NewUsernameChangedNotifier(cmr chatmember.Repository) *UsernameChangedNotifier {
	return &UsernameChangedNotifier{
		chatMemberRepository: cmr,
	}
}

func (n *UsernameChangedNotifier) NotifyUsernameChanged(c *botapi.Context, oldUsername, newUsername string) error {
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	loc := cctx.MustLocalizer(c)
	args := i18n.SystemUsernameChangedMaleData{
		User:        tghtml.MemberMention(loc, ch, cm),
		OldUsername: tghtml.Code("@" + oldUsername),
		NewUsername: tghtml.Code("@" + newUsername),
	}

	text := loc.TGender(cm.User.Gender, i18n.System.UsernameChanged, args)

	admins, err := n.chatMemberRepository.ListAdmins(c, ch.ID, chatmember.StatusModerator)
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
