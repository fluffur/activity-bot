package model

import "time"

type Message struct {
	ChatID int64
	UserID int64
}

type MessageActivity struct {
	Count int64
	Date  time.Time
}

func NewMessage(chatID int64, userID int64) Message {
	return Message{
		ChatID: chatID,
		UserID: userID,
	}
}
