package adapter

import (
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type TelegramModerator struct {
	bot *gotgbot.Bot
}

func NewTelegramModerator(b *gotgbot.Bot) *TelegramModerator {
	return &TelegramModerator{b}
}

func (m *TelegramModerator) Kick(chatID, userID int64) error {
	_, err := m.bot.BanChatMember(chatID, userID, &gotgbot.BanChatMemberOpts{})
	if err != nil {
		return err
	}
	_, err = m.bot.UnbanChatMember(chatID, userID, &gotgbot.UnbanChatMemberOpts{})
	return err
}

func (m *TelegramModerator) Ban(chatID, userID int64, untilDate time.Time) error {
	opts := &gotgbot.BanChatMemberOpts{}
	if !untilDate.IsZero() {
		opts.UntilDate = untilDate.Unix()
	}
	_, err := m.bot.BanChatMember(chatID, userID, opts)
	return err
}

func (m *TelegramModerator) Mute(chatID, userID int64, untilDate time.Time) error {
	permissions := gotgbot.ChatPermissions{
		CanSendMessages:       false,
		CanSendAudios:         false,
		CanSendDocuments:      false,
		CanSendPhotos:         false,
		CanSendVideos:         false,
		CanSendVideoNotes:     false,
		CanSendVoiceNotes:     false,
		CanSendPolls:          false,
		CanSendOtherMessages:  false,
		CanAddWebPagePreviews: false,
	}
	opts := &gotgbot.RestrictChatMemberOpts{}
	if !untilDate.IsZero() {
		opts.UntilDate = untilDate.Unix()
	}
	_, err := m.bot.RestrictChatMember(chatID, userID, permissions, opts)
	return err
}

func (m *TelegramModerator) Unban(chatID, userID int64) error {
	_, err := m.bot.UnbanChatMember(chatID, userID, nil)
	return err
}

func (m *TelegramModerator) Unmute(chatID, userID int64) error {
	permissions := gotgbot.ChatPermissions{
		CanSendMessages:       true,
		CanSendAudios:         true,
		CanSendDocuments:      true,
		CanSendPhotos:         true,
		CanSendVideos:         true,
		CanSendVideoNotes:     true,
		CanSendVoiceNotes:     true,
		CanSendPolls:          true,
		CanSendOtherMessages:  true,
		CanAddWebPagePreviews: true,
	}
	_, err := m.bot.RestrictChatMember(chatID, userID, permissions, nil)
	return err
}
