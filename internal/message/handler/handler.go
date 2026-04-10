package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/member"
	"activity-bot/internal/message"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service       *message.Service
	memberService *member.Service
	chatService   *chat.Service
}

func New(service *message.Service, memberService *member.Service, chatService *chat.Service) *Handler {
	return &Handler{service, memberService, chatService}
}

func (h *Handler) Bot(b *gotgbot.Bot, ctx *command.Context) error {
	//c, err := ctx.Chat()
	//if err != nil {
	//	return err
	//}
	//request := &deepseek.ChatCompletionRequest{
	//	Model: deepseek.DeepSeekChat,
	//	Messages: []deepseek.ChatCompletionMessage{
	//		{
	//			Role:    deepseek.ChatMessageRoleSystem,
	//			Content: "Отвечай коротко, 1-2 предложения. " + c.AISystemPrompt,
	//		},
	//		{
	//			Role:    deepseek.ChatMessageRoleUser,
	//			Content: ctx.TextOrDefault(""),
	//		},
	//	},
	//}
	//resp, err := h.deepseekClient.CreateChatCompletion(ctx.StdContext(), request)
	//if err != nil {
	//	_ = ctx.Reply(b, "Не удалось получить ответ от бота", nil)
	//	return err
	//}
	//
	//if len(resp.Choices) == 0 {
	//	_ = ctx.Reply(b, "Бот не вернул ответ", nil)
	//	return nil
	//}
	//
	//return ctx.Reply(b, resp.Choices[0].Message.Content, &gotgbot.SendMessageOpts{
	//	ParseMode: gotgbot.ParseModeMarkdown,
	//})
	return nil
}

func (h *Handler) Message(_ *gotgbot.Bot, ctx *command.Context) error {
	if ctx.EffectiveMessage.ChatOwnerLeft != nil || ctx.EffectiveMessage.LeftChatMember != nil || len(ctx.EffectiveMessage.NewChatMembers) > 0 {
		return nil
	}
	effectiveSender := ctx.EffectiveSender
	effectiveChat := ctx.EffectiveChat
	if effectiveSender == nil || effectiveChat == nil {
		return nil
	}
	if _, err := h.memberService.EnsureMemberExists(
		ctx.StdContext(),
		effectiveChat.Id,
		effectiveSender.Id(),
		effectiveSender.Username(),
		effectiveSender.FirstName(),
		effectiveSender.LastName(),
		ctx.EffectiveMessage.SenderTag,
	); err != nil {
		return err
	}

	return h.service.Save(
		ctx.StdContext(),
		effectiveChat.Id,
		effectiveSender.Id(),
		ctx.EffectiveMessage.MessageId,
	)
}
