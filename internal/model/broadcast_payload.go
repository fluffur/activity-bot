package model

type BroadcastPayload struct {
	ChatID     int64 `json:"chat_id"`
	FromChatID int64 `json:"from_chat_id"`
	MessageID  int64 `json:"message_id"`
}
