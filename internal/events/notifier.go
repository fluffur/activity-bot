package events

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/summon"
	"activity-bot/internal/utils/tghtml"
	"fmt"

	"github.com/gotd/botapi"
)

type UsernameChangedNotifier struct {
	chatMemberRepository chatmember.Repository
	summonH              *summon.Handler
}

func NewUsernameChangedNotifier(cmr chatmember.Repository, summonH *summon.Handler) *UsernameChangedNotifier {
	return &UsernameChangedNotifier{
		chatMemberRepository: cmr,
		summonH:              summonH,
	}
}

func (n *UsernameChangedNotifier) NotifyUsernameChanged(c *botapi.Context, oldUsername, newUsername string) error {
	msg := c.Message()
	if msg == nil {
		return nil
	}
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	loc := cctx.MustLocalizer(c)

	admins, err := n.chatMemberRepository.ListAdmins(c, ch.ID, ch.UsernameChangedNotifyStatus)
	if err != nil {
		return fmt.Errorf("notify username changed list admins: %w", err)
	}

	if len(admins) == 0 {
		return nil
	}

	args := i18n.SystemUsernameChangedData{
		User:        tghtml.MemberMention(loc, ch, cm),
		OldUsername: tghtml.Code("@" + oldUsername),
		NewUsername: tghtml.Code("@" + newUsername),
	}

	var locKey i18n.MessageID
	if oldUsername == "" {
		locKey = i18n.System.UsernameAdded
	}
	if newUsername == "" {
		locKey = i18n.System.UsernameDeleted
	}
	if newUsername != "" && oldUsername != newUsername {
		locKey = i18n.System.UsernameChanged
	}

	text := loc.T(
		locKey,
		args,
		i18n.WithGender(cm.Gender()),
	)

	return n.summonH.Summon(
		c,
		text,
		msg.MessageID,
		ch,
		admins,
	)
}
