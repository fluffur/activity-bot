package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/model"
	"encoding/json"
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/hibiken/asynq"
)

type Handler struct {
	chatService *chat.Service
	asyncClient *asynq.Client
}

func New(chatService *chat.Service, asyncClient *asynq.Client) *Handler {
	return &Handler{chatService, asyncClient}
}

func (h *Handler) Post(b *gotgbot.Bot, ctx *cmd.Context) error {
	log.Println("post")
	if ctx.EffectiveChat.Id != -1003824019217 {
		return nil
	}

	chats, err := h.chatService.GetChatsWithEnabledBroadcast(ctx.StdContext())
	if err != nil {
		return fmt.Errorf("failed to get chats: %w", err)
	}
	msg := ctx.EffectiveMessage

	for _, c := range chats {

		payload, _ := json.Marshal(model.BroadcastPayload{
			ChatID:     c.ID,
			FromChatID: msg.Chat.Id,
			MessageID:  msg.MessageId,
		})

		task := asynq.NewTask("broadcast:post", payload)

		_, err := h.asyncClient.Enqueue(task)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) Unsubscribe(b *gotgbot.Bot, ctx *cmd.Context) error {

	chatID := ctx.EffectiveChat.Id

	err := h.chatService.DisableBroadcast(ctx.StdContext(), chatID)
	if err != nil {
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditReplyMarkup(b, nil)
	if err != nil {
		return err
	}

	_, err = b.SendMessage(chatID, "Рассылка отключена ❌", nil)

	return err
}
