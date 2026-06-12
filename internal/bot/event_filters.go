package bot

import (
	"log"

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
	c := u.EffectiveChat()

	msg := u.EffectiveMessage
	if c.GetID() == -1003672433876 {
		log.Println(u)
	}
	if msg == nil || msg.Action == nil {
		return false
	}
	if c.GetID() == -1003672433876 {
		log.Println("action", msg.Action)
	}

	_, ok := msg.Action.(*tg.MessageActionChatDeleteUser)
	return ok
}
