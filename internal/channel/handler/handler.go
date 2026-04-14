package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"encoding/json"
	"fmt"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/hibiken/asynq"
)

type Handler struct {
	memberService *member.Service
	chatService   *chat.Service
	asyncClient   *asynq.Client
	channelID     int64
}

func New(memberService *member.Service, chatService *chat.Service, asyncClient *asynq.Client, channelID int64) *Handler {
	return &Handler{memberService, chatService, asyncClient, channelID}
}

func (h *Handler) Post(ctx *command.Context, u *ext.Update) error {
	if u.EffectiveChat().GetID() != h.channelID {
		return nil
	}

	chats, err := h.chatService.GetChatsWithEnabledBroadcast(ctx.StdContext())
	if err != nil {
		return fmt.Errorf("failed to get chats: %w", err)
	}
	msg := u.EffectiveMessage

	for _, c := range chats {

		payload, _ := json.Marshal(model.BroadcastPayload{
			ChatID:     c.ID,
			FromChatID: u.EffectiveChat().GetID(),
			MessageID:  int64(msg.ID),
		})

		task := asynq.NewTask("broadcast:post", payload)

		_, err := h.asyncClient.Enqueue(task)
		if err != nil {
			return fmt.Errorf("post: enqueue broadcast task: %w", err)
		}
	}

	return nil
}

func (h *Handler) Unsubscribe(ctx *command.Context, u *ext.Update) error {
	effectiveChat := u.EffectiveChat()
	effectiveSender := u.EffectiveUser()

	m, err := h.memberService.GetChatMember(ctx.StdContext(), effectiveChat.GetID(), effectiveSender.GetID())
	if err != nil {
		return fmt.Errorf("unsubscribe: get member: %w", err)
	}
	if !m.StatusGranted(model.StatusCoOwner) {
		_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			Message: "У вас нет прав адмниистратора для этого",
		})
		return fmt.Errorf("unsubscribe: answer callback: %w", err)
	}
	err = h.chatService.DisableBroadcast(ctx.StdContext(), effectiveChat.GetID())
	if err != nil {
		return fmt.Errorf("unsubscribe: disable broadcast: %w", err)
	}

	_, err = ctx.EditMessage(effectiveChat.GetID(), &tg.MessagesEditMessageRequest{
		ID:          u.CallbackQuery.GetMsgID(),
		ReplyMarkup: nil,
	})
	if err != nil {
		return fmt.Errorf("unsubscribe: edit message: %w", err)
	}

	return ctx.ReplyOnly(u, options.WithText("Рассылка отключена ❌"))
}

func (h *Handler) Subscribe(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return fmt.Errorf("subscribe: get chat: %w", err)
	}
	if err := h.chatService.EnableBroadcast(ctx.StdContext(), c.ID); err != nil {
		return fmt.Errorf("subscribe: enable broadcast: %w", err)
	}
	return ctx.ReplyOnly(u, options.WithText("Рассылка включена"))
}
