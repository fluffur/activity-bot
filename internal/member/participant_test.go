package member

import (
	"testing"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func TestLeaveFromUpdate_ChannelParticipantLeft(t *testing.T) {
	u := &ext.Update{
		UpdateClass: &tg.UpdateChannelParticipant{
			ChannelID: 1234567890,
			UserID:    42,
			NewParticipant: &tg.ChannelParticipantLeft{
				Peer: &tg.PeerUser{UserID: 42},
			},
			PrevParticipant: &tg.ChannelParticipant{UserID: 42},
		},
		ChannelParticipant: &tg.UpdateChannelParticipant{
			ChannelID: 1234567890,
			UserID:    42,
			NewParticipant: &tg.ChannelParticipantLeft{
				Peer: &tg.PeerUser{UserID: 42},
			},
			PrevParticipant: &tg.ChannelParticipant{UserID: 42},
		},
	}

	chatID, userID, ok := LeaveFromUpdate(u)
	if !ok {
		t.Fatal("expected leave update")
	}
	if userID != 42 {
		t.Fatalf("userID = %d, want 42", userID)
	}
	if chatID != -1001234567890 {
		t.Fatalf("chatID = %d, want -1001234567890", chatID)
	}
}

func TestLeaveFromUpdate_ChatParticipantDelete(t *testing.T) {
	u := &ext.Update{
		UpdateClass: &tg.UpdateChatParticipantDelete{
			ChatID: 100,
			UserID: 7,
		},
	}

	chatID, userID, ok := LeaveFromUpdate(u)
	if !ok {
		t.Fatal("expected leave update")
	}
	if userID != 7 {
		t.Fatalf("userID = %d, want 7", userID)
	}
	if chatID != -100 {
		t.Fatalf("chatID = %d, want -100", chatID)
	}
}
