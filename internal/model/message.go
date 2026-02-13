package model

import "time"

type Message struct {
	MessageID int64
	ChatID    int64
	UserID    int64
}

type MessageActivity struct {
	Count int64
	Date  time.Time
}

func NewMessage(chatID, userID, id int64) Message {
	return Message{
		MessageID: id,
		ChatID:    chatID,
		UserID:    userID,
	}
}
