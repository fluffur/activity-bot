package adapter

import "github.com/PaulSonOfLars/gotgbot/v2"

type TelegramMemberStatusProvider struct {
	bot *gotgbot.Bot
}

func NewTelegramMemberStatusProvider(b *gotgbot.Bot) *TelegramMemberStatusProvider {
	return &TelegramMemberStatusProvider{b}
}

func (p *TelegramMemberStatusProvider) GetChatMemberStatus(chatID, userID int64) (string, error) {
	member, err := p.bot.GetChatMember(chatID, userID, nil)
	if err != nil {
		return "", err
	}
	return member.GetStatus(), nil
}
