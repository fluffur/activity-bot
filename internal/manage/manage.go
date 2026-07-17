package manage

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/i18n"
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

	loc := cctx.MustLocalizer(c)

	sender := c.Sender()
	if sender == nil {
		return fmt.Errorf("manage: sender is nil")
	}

	senderID := sender.ID
	search, _ := cctx.MustArgs(c).Text()

	chats, err := h.chatService.GetUserManagedChats(c, senderID, search)
	if err != nil {
		return fmt.Errorf("get managed chats: %w", err)
	}

	if len(chats) == 0 {
		_, err := c.Reply(loc.T(i18n.Cmd.Manage.NoChats, nil))
		return err
	}

	_, err = c.Reply(
		loc.T(i18n.Cmd.Manage.Select, nil),
		botapi.WithReplyMarkup(manageKeyboard(loc, chats)),
	)

	return err
}

func manageKeyboard(loc *i18n.Localizer, chats []chat.Chat) *botapi.InlineKeyboardMarkup {
	rows := make([][]botapi.InlineKeyboardButton, 0, len(chats))

	for _, ch := range chats {
		title := ch.Title
		if title == "" {
			title = loc.T(i18n.Cmd.Manage.Untitled, nil)
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
	loc := cctx.MustLocalizer(c)

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
			botapi.WithCallbackText(loc.T(i18n.Cmd.Manage.NotFound, nil)),
		); err != nil {
			return err
		}

		return fmt.Errorf("callback manage: get chat %d: %w", chatID, err)
	}

	if err := h.sessionRepository.SetSession(c, senderID, chatID); err != nil {
		return fmt.Errorf("callback manage: set session: %w", err)
	}

	if err := c.AnswerCallback(
		botapi.WithCallbackText(loc.T(i18n.Cmd.Manage.Selected, nil)),
	); err != nil {
		return err
	}

	text := loc.T(
		i18n.Cmd.Manage.Success,
		i18n.CmdManageSuccessData{
			Chat: tghtml.Bold(tghtml.Escape(selected.Title)),
		},
	)

	_, err = c.Bot.EditMessageText(
		c,
		botapi.ID(cq.Message.Chat.ID),
		cq.Message.MessageID,
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}
