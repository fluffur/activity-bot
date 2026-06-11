package member

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func LeaveFromUpdate(u *ext.Update) (chatID, userID int64, ok bool) {
	chat := u.EffectiveChat()
	if chat == nil || chat.GetID() == 0 {
		return 0, 0, false
	}
	chatID = chat.GetID()

	if cp := u.ChannelParticipant; cp != nil {
		switch cp.NewParticipant.(type) {
		case *tg.ChannelParticipantLeft, *tg.ChannelParticipantBanned:
			return chatID, cp.UserID, true
		}
		return 0, 0, false
	}

	if cp := u.ChatParticipant; cp != nil {
		if cp.NewParticipant == nil && cp.PrevParticipant != nil {
			return chatID, cp.UserID, true
		}
	}

	return 0, 0, false
}
