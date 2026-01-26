package adapter

import "github.com/PaulSonOfLars/gotgbot/v2"

type TelegramChatMemberProvider struct {
	bot *gotgbot.Bot
}

func (p *TelegramChatMemberProvider) GetChatMemberStatus(chatID, userID int64) (string, error) {
	member, err := p.bot.GetChatMember(chatID, userID, nil)
	if err != nil {
		return "", err
	}
	return member.GetStatus(), nil
}
