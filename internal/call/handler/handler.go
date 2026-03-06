package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/model"
	"activity-bot/internal/session"
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service        *call.Service
	chatService    *chat.Service
	adminService   *admin.Service
	sessionService *session.Service
}

func New(service *call.Service, chatService *chat.Service, adminService *admin.Service, sessionService *session.Service) *Handler {
	return &Handler{service, chatService, adminService, sessionService}
}

func (h *Handler) Call(b *gotgbot.Bot, ctx *cmd.Context) error {
	members, err := h.service.GetAllMembers(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return fmt.Errorf("failed to get chat members: %w", err)
	}
	if len(members) == 0 {
		return ctx.Reply(b, "Не найдено пользователей для созыва, скорее всего бот был добавлен недавно и понадобится время, чтобы он успел познакомиться со всеми участниками!", nil)
	}
	return h.handleCall(b, ctx, members)
}

func (h *Handler) CallInactive(b *gotgbot.Bot, ctx *cmd.Context) error {
	members, err := h.service.GetInactiveMembers(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return fmt.Errorf("failed to get inactive members: %w", err)
	}
	if len(members) == 0 {
		return ctx.Reply(b, "Нет участников, не писавших более суток", nil)
	}
	return h.handleCall(b, ctx, members)
}

func (h *Handler) CallInactiveCallback(b *gotgbot.Bot, ctx *cmd.Context) error {

	isAdmin, err := h.adminService.IsAdmin(ctx.StdContext(), ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
	if err != nil {
		return fmt.Errorf("failed to check admin rights: %w", err)
	}
	if !isAdmin {
		_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Требуются права администратора",
		})
		return err
	}
	_, _ = ctx.CallbackQuery.Answer(b, nil)

	return h.CallInactive(b, ctx)
}

func (h *Handler) handleCall(b *gotgbot.Bot, ctx *cmd.Context, members []model.ChatMember) error {
	var replyParams *gotgbot.ReplyParameters
	if ctx.EffectiveMessage.ReplyToMessage != nil {
		replyParams = &gotgbot.ReplyParameters{
			ChatId:    ctx.EffectiveChat.Id,
			MessageId: ctx.EffectiveMessage.ReplyToMessage.MessageId,
		}
	}

	chatSettings, err := h.service.GetChatSettings(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	mentionsLimit := int(chatSettings.MentionsPerMessage)
	if mentionsLimit <= 0 {
		mentionsLimit = 5
	}

	message := ctx.HTML()
	if message == "" {
		message = chatSettings.WelcomeCallMessage
	}

	if message != "" {
		message = view.ReplaceMentionsWithLinks(message)
	}

	for i := 0; i < len(members); i += mentionsLimit {
		end := i + mentionsLimit
		if end > len(members) {
			end = len(members)
		}

		chunkText := view.FormatCallChunk(message, members[i:end], chatSettings.MentionTypes)

		if len(ctx.EffectiveMessage.Photo) > 0 {
			lastPhoto := ctx.EffectiveMessage.Photo[len(ctx.EffectiveMessage.Photo)-1]
			if _, err := b.SendPhoto(ctx.TargetChatID(), gotgbot.InputFileByID(lastPhoto.FileId), &gotgbot.SendPhotoOpts{
				ParseMode:       gotgbot.ParseModeHTML,
				Caption:         chunkText,
				HasSpoiler:      ctx.EffectiveMessage.HasMediaSpoiler,
				ReplyParameters: replyParams,
			}); err != nil {
				return err
			}
		} else {
			if _, err := ctx.EffectiveMessage.Reply(b, chunkText, &gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyParameters: replyParams,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Handler) SetMentionsPerMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	countStr := ctx.FirstArgument()
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || count > 100 {
		return ctx.Reply(b, "Укажите число от 1 до 100", nil)
	}

	if err := h.service.SetMentionsPerMessage(ctx.StdContext(), ctx.TargetChatID(), int32(count)); err != nil {
		return err
	}

	return ctx.Reply(b, fmt.Sprintf("Лимит упоминаний в одном сообщении изменен на %d", count), nil)
}

func (h *Handler) ShowCallTypes(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.Reply(b, "Выберите типы упоминаний для команды call:", &gotgbot.SendMessageOpts{
		ReplyMarkup: h.getCallTypesKeyboard(int32(c.MentionTypes)),
	})
}

func (h *Handler) CallbackCallType(b *gotgbot.Bot, ctx *cmd.Context) error {
	chatID := ctx.TargetChatID()
	isAdmin, err := h.adminService.IsAdmin(ctx.StdContext(), chatID, ctx.EffectiveSender.Id())
	if err != nil {
		return err
	}
	if !isAdmin {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "У вас нет прав администратора для выполнения действия",
		})
		return err
	}
	var bit int32
	if _, err := fmt.Sscanf(ctx.CallbackQuery.Data, "call_type:%d", &bit); err != nil {
		return err
	}

	c, err := h.chatService.GetChat(ctx.StdContext(), chatID)
	if err != nil {
		return err
	}

	current := c.MentionTypes
	var newTypes int32

	if bit == view.MentionTypeNWSP {
		newTypes = view.MentionTypeNWSP
	} else {
		current &^= view.MentionTypeNWSP
		newTypes = current ^ bit
	}
	if newTypes == c.MentionTypes {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Настройки не изменились",
		})
		return err
	}
	if err := h.service.SetMentionTypes(ctx.StdContext(), chatID, newTypes); err != nil {
		return err
	}

	_, _, err = ctx.EffectiveMessage.EditReplyMarkup(b, &gotgbot.EditMessageReplyMarkupOpts{
		ReplyMarkup: h.getCallTypesKeyboard(newTypes),
	})
	if err != nil {
		return err
	}

	_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Настройки обновлены"})
	return err
}

func (h *Handler) getCallTypesKeyboard(currentTypes int32) gotgbot.InlineKeyboardMarkup {
	types := []struct {
		name string
		bit  int32
	}{
		{"Пустота", view.MentionTypeNWSP},
		{"Эмодзи", view.MentionTypeEmoji},
		{"Имя", view.MentionTypeName},
		{"Роль", view.MentionTypeRole},
	}

	var rows [][]gotgbot.InlineKeyboardButton
	var row []gotgbot.InlineKeyboardButton

	for i, t := range types {
		status := ""
		checkMark := ""

		if currentTypes&t.bit > 0 {
			status = "primary"
			checkMark = "✅ "
		}

		btn := gotgbot.InlineKeyboardButton{
			Text:         fmt.Sprintf("%s%s", checkMark, t.name),
			CallbackData: fmt.Sprintf("call_type:%d", t.bit),
			Style:        status,
		}

		row = append(row, btn)

		if (i+1)%2 == 0 {
			rows = append(rows, row)
			row = nil
		}
	}

	if len(row) > 0 {
		rows = append(rows, row)
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
func (h *Handler) SetWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	message := ctx.FirstArgument()
	if err := h.service.SetWelcomeCallMessage(ctx.StdContext(), ctx.TargetChatID(), message); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatWelcomeCallMessageSet(), nil)
}

func (h *Handler) EnableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.EnableCallOnJoin(ctx.StdContext(), ctx.TargetChatID()); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinEnabled(), nil)
}

func (h *Handler) DisableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.DisableCallOnJoin(ctx.StdContext(), ctx.TargetChatID()); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinDisabled(), nil)
}

func (h *Handler) ShowWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}
	return ctx.Reply(b, view.FormatWelcomeCallMessage(c.WelcomeCallMessage), nil)
}
