package model

type Message struct {
	ChatID int64
	UserID int64
}

func NewMessage(chatID int64, userID int64) Message {
	return Message{
		ChatID: chatID,
		UserID: userID,
	}
}
