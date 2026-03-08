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
	channelID   int64
}

func New(chatService *chat.Service, asyncClient *asynq.Client, channelID int64) *Handler {
	return &Handler{chatService, asyncClient, channelID}
}

func (h *Handler) Post(b *gotgbot.Bot, ctx *cmd.Context) error {
	log.Println("post", ctx.EffectiveChat.Id, h.channelID)
	if ctx.EffectiveChat.Id != h.channelID {
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

	return ctx.Reply(b, "Рассылка отключена ❌", nil)
}

func (h *Handler) Subscribe(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.EffectiveChat.Id
	if err := h.chatService.EnableBroadcast(ctx.StdContext(), chatID); err != nil {
		return err
	}
	return ctx.Reply(b, "Рассылка включена", nil)
}
