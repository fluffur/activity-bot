package bot

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func textMessageFilter(u *ext.Update) bool {
	return u.EffectiveMessage != nil && u.EffectiveMessage.Text != ""
}

func chatTitleChangedFilter(u *ext.Update) bool {
	msg := u.EffectiveMessage
	if msg == nil || msg.Action == nil {
		return false
	}

	_, ok := msg.Action.(*tg.MessageActionChatEditTitle)
	return ok
}

func joinMemberFilter(u *ext.Update) bool {
	msg := u.EffectiveMessage
	if msg == nil || msg.Action == nil {
		return false
	}

	switch msg.Action.(type) {
	case *tg.MessageActionChatAddUser, *tg.MessageActionChatJoinedByLink, *tg.MessageActionChatJoinedByRequest:
		return true
	default:
		return false
	}
}

func leftMemberFilter(u *ext.Update) bool {
	msg := u.EffectiveMessage
	if msg == nil || msg.Action == nil {
		return false
	}

	_, ok := msg.Action.(*tg.MessageActionChatDeleteUser)
	return ok
}
