package moderation

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/botapi"
)

func (h *Handler) Manage(c *botapi.Context) error {
	msg := c.Message()
	if msg == nil {
		return fmt.Errorf("manage: msg is nil")
	}
	senderID := msg.Chat.ID
	search, _ := cctx.MustArgs(c).Text()

	chats, err := h.service.GetUserManagedChats(c, senderID, search)
	if err != nil {
		return fmt.Errorf("get managed chats: %w", err)
	}

	if len(chats) == 0 {
		_, err := c.Reply("❌ У вас нет доступных чатов для управления.")
		return err
	}

	_, err = c.Reply(
		"Выберите чат для управления:",
		botapi.WithReplyMarkup(manageKeyboard(chats)),
	)

	return err
}

func manageKeyboard(chats []chat.Chat) *botapi.InlineKeyboardMarkup {
	rows := make([][]botapi.InlineKeyboardButton, 0, len(chats))

	for _, ch := range chats {
		title := ch.Title
		if title == "" {
			title = "Чат без названия"
		}

		rows = append(rows, []botapi.InlineKeyboardButton{
			{
				Text:         title,
				CallbackData: fmt.Sprintf("manage:%d", ch.ID),
			},
		})
	}

	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (h *Handler) CallbackManage(c *botapi.Context) error {

	cq := c.Update.CallbackQuery
	if cq == nil {
		return fmt.Errorf("callback query is nil")
	}
	data := cq.Data
	senderID := cq.From.ID
	chatIDStr := strings.TrimPrefix(data, "manage:")

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("parse chat id: %w", err)
	}

	selected, err := h.chatService.Get(c, chatID)
	if err != nil {
		if err := c.AnswerCallback(
			botapi.WithCallbackText("Выбранный чат не найден"),
		); err != nil {
			return err
		}
		return fmt.Errorf("callback manage: get chat %d: %w", chatID, err)
	}

	if err := h.sessionRepository.SetSession(c, senderID, chatID); err != nil {
		return fmt.Errorf("callback manage: set session: %w", err)
	}

	if err := c.AnswerCallback(
		botapi.WithCallbackText("Чат выбран."),
	); err != nil {
		return err
	}

	_, err = c.Bot.EditMessageText(
		c,
		botapi.ID(cq.Message.Chat.ID),
		cq.Message.MessageID,
		fmt.Sprintf(
			"✅ Теперь вы управляете чатом:\n<b>%s</b>\n\nВсе команды настроек будут применяться к нему.",
			tghtml.Escape(selected.Title),
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}
