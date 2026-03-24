package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/member"
	"activity-bot/internal/model"

	"github.com/PaulSonOfLars/gotgbot/v2"
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

func (h *Handler) Post(b *gotgbot.Bot, ctx *cmd.Context) error {
	//log.Println("post", ctx.EffectiveChat.Id, h.channelID)
	//if ctx.EffectiveChat.Id != h.channelID {
	//	return nil
	//}
	//
	//chats, err := h.chatService.GetChatsWithEnabledBroadcast(ctx.StdContext())
	//if err != nil {
	//	return fmt.Errorf("failed to get chats: %w", err)
	//}
	//msg := ctx.EffectiveMessage
	//
	//for _, c := range chats {
	//
	//	payload, _ := json.Marshal(model.BroadcastPayload{
	//		ChatID:     c.ID,
	//		FromChatID: msg.Chat.Id,
	//		MessageID:  msg.MessageId,
	//	})
	//
	//	task := asynq.NewTask("broadcast:post", payload)
	//
	//	_, err := h.asyncClient.Enqueue(task)
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	return nil
}

func (h *Handler) Unsubscribe(b *gotgbot.Bot, ctx *cmd.Context) error {

	m, err := h.memberService.GetChatMember(ctx.StdContext(), ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}
	if !m.StatusGranted(model.StatusCoOwner) {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "У вас нет прав адмниистратора для этого",
		})
		return err
	}
	err = h.chatService.DisableBroadcast(ctx.StdContext(), ctx.EffectiveChat.Id)
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
