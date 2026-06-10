package member

import (
	"activity-bot/internal/helpers"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func isChannelParticipantLeft(p tg.ChannelParticipantClass) bool {
	switch v := p.(type) {
	case *tg.ChannelParticipantLeft:
		return true
	case *tg.ChannelParticipantBanned:
		return v.Left
	default:
		return false
	}
}

func LeaveFromUpdate(u *ext.Update) (chatID, userID int64, ok bool) {
	if upd, isDelete := u.UpdateClass.(*tg.UpdateChatParticipantDelete); isDelete {
		return helpers.BasicChatID(upd.ChatID), upd.UserID, true
	}

	if cp := u.ChannelParticipant; cp != nil {
		if cp.UserID == 0 {
			return 0, 0, false
		}
		if isChannelParticipantLeft(cp.NewParticipant) {
			return helpers.ChannelChatID(cp.ChannelID), cp.UserID, true
		}
		if cp.NewParticipant == nil && cp.PrevParticipant != nil {
			return helpers.ChannelChatID(cp.ChannelID), cp.UserID, true
		}
		return 0, 0, false
	}

	if cp := u.ChatParticipant; cp != nil {
		if cp.NewParticipant == nil && cp.PrevParticipant != nil {
			return helpers.BasicChatID(cp.ChatID), cp.UserID, true
		}
	}

	return 0, 0, false
}
