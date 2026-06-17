package participant

import (
	"github.com/gotd/td/tg"
)

func Rank(p tg.ChannelParticipantClass) string {
	if p == nil {
		return ""
	}

	switch v := p.(type) {
	case *tg.ChannelParticipant:
		return v.Rank
	case *tg.ChannelParticipantAdmin:
		return v.Rank
	case *tg.ChannelParticipantCreator:
		return v.Rank
	case *tg.ChannelParticipantSelf:
		if rank, ok := v.GetRank(); ok {
			return rank
		}
	}

	return ""
}

func ID(p tg.ChannelParticipantClass) int64 {
	if p == nil {
		return 0
	}

	switch v := p.(type) {
	case *tg.ChannelParticipant:
		return v.UserID
	case *tg.ChannelParticipantAdmin:
		return v.UserID
	case *tg.ChannelParticipantCreator:
		return v.UserID
	case *tg.ChannelParticipantSelf:
		return v.UserID
	}

	return 0
}
