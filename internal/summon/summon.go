package summon

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) SummonAll(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("summon chat: %w", err)
	}

	m := c.Message()
	if m == nil {
		return nil
	}

	args := predicate.Args(c)

	if !ch.SkipCallConfirmation {
		return h.RequestSummonConfirmation(c, TextFromArgs(ch, args), m.MessageID)
	}

	cms, err := h.chatMemberService.ListSummonChatMembers(c.Context, ch.ID)
	if err != nil {
		return fmt.Errorf("summon cms list: %w", err)
	}

	return h.Summon(c, TextFromArgs(ch, args), m.MessageID, ch, cms)
}

func (h *Handler) RequestSummonConfirmation(
	c *botapi.Context,
	text string,
	msgID int,
) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return err
	}

	cm, err := cctx.ChatMember(c.Context)
	if err != nil {
		return err
	}

	err = h.summonFSM.Enter(c, StateAwaitConfirmation, StateData{
		Text:      text,
		MessageID: msgID,
		ChatID:    ch.ID,
		UserID:    cm.User.ID,
	})
	if err != nil {
		return err
	}

	_, err = c.Reply(
		h.translator.T(ch.Lang, i18n.Cmd.Summon.Confirm.Text),
		botapi.WithReplyMarkup(
			botapi.InlineKeyboard(
				botapi.InlineRow(
					botapi.InlineButtonData(
						h.translator.T(ch.Lang, i18n.Cmd.Summon.Confirm.Yes),
						"summon:confirm",
					),
					botapi.InlineButtonData(
						h.translator.T(ch.Lang, i18n.Cmd.Summon.Confirm.No),
						"summon:cancel",
					),
				),
				botapi.InlineRow(
					botapi.InlineButtonData(
						h.translator.T(
							ch.Lang,
							i18n.Cmd.Summon.Confirm.YesAndDisable,
						),
						"summon:confirm_dont_ask",
					),
				),
			),
		),
	)

	return err
}

func (h *Handler) ConfirmSummon(c *botapi.Context) error {
	data, ch, err := h.summonSession(c)
	if err != nil || data == nil {
		return err
	}

	if err := h.summonFSM.Clear(c); err != nil {
		return err
	}

	_ = c.AnswerCallback()

	chatID, _ := c.Chat()
	_ = c.Bot.DeleteMessage(
		c.Context,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
	)

	cms, err := h.chatMemberService.ListSummonChatMembers(
		c.Context,
		ch.ID,
	)
	if err != nil {
		return fmt.Errorf("summon cms list: %w", err)
	}

	return h.Summon(
		c,
		data.Text,
		data.MessageID,
		ch,
		cms,
	)
}

func (h *Handler) CancelSummon(c *botapi.Context) error {
	data, ch, err := h.summonSession(c)
	if err != nil || data == nil {
		return err
	}

	if err := h.summonFSM.Clear(c); err != nil {
		return err
	}

	chatID, _ := c.Chat()

	_, _ = c.Bot.EditMessageText(
		c.Context,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		h.translator.T(
			ch.Lang,
			i18n.Cmd.Summon.Confirm.Canceled,
		),
	)

	return c.AnswerCallback(
		botapi.WithCallbackText(
			h.translator.T(
				ch.Lang,
				i18n.Cmd.Summon.Confirm.Canceled,
			),
		),
	)
}

func (h *Handler) ConfirmSummonDontAsk(c *botapi.Context) error {
	data, ch, err := h.summonSession(c)
	if err != nil || data == nil {
		return err
	}

	if err := h.chatService.SetSkipSummonConfirmation(
		c.Context,
		ch.ID,
		true,
	); err != nil {
		return fmt.Errorf("update chat: %w", err)
	}

	if err := h.summonFSM.Clear(c); err != nil {
		return err
	}

	_ = c.AnswerCallback(
		botapi.WithCallbackText(
			h.translator.T(
				ch.Lang,
				i18n.Cmd.Summon.Confirm.Disabled,
			),
		),
	)

	chatID, _ := c.Chat()

	_, _ = c.Bot.EditMessageText(
		c.Context,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		h.translator.T(
			ch.Lang,
			i18n.Cmd.Summon.Confirm.Disabled,
		),
	)

	cms, err := h.chatMemberService.ListSummonChatMembers(
		c.Context,
		ch.ID,
	)
	if err != nil {
		return fmt.Errorf("summon cms list: %w", err)
	}

	return h.Summon(
		c,
		data.Text,
		data.MessageID,
		ch,
		cms,
	)
}

func (h *Handler) SummonStyle(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("summon style chat: %w", err)
	}

	_, err = c.Reply(
		h.translator.T(ch.Lang, i18n.Cmd.Summon.Style.Text),
		botapi.WithReplyMarkup(mentionTypesKeyboard(h.translator, ch.Lang, ch.MentionTypes)),
	)

	return err
}

func (h *Handler) ToggleMentionEmoji(c *botapi.Context) error {
	return h.toggleMentionStyle(c, chat.MentionEmoji)
}

func (h *Handler) ToggleMentionName(c *botapi.Context) error {
	return h.toggleMentionStyle(c, chat.MentionName)
}

func (h *Handler) ToggleMentionRole(c *botapi.Context) error {
	return h.toggleMentionStyle(c, chat.MentionRole)
}

func (h *Handler) toggleMentionStyle(c *botapi.Context, flag chat.MentionTypes) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("toggle mention type: %w", err)
	}

	if ch.MentionTypes.Has(flag) {
		ch.MentionTypes.Remove(flag)
	} else {
		ch.MentionTypes.Add(flag)
	}

	if err := h.chatService.SetMentionTypes(
		c.Context,
		ch.ID,
		ch.MentionTypes,
	); err != nil {
		return err
	}

	_, _ = c.Bot.EditMessageReplyMarkup(
		c.Context,
		botapi.ID(ch.ID),
		c.Update.CallbackQuery.Message.MessageID,
		mentionTypesKeyboard(
			h.translator,
			ch.Lang,
			ch.MentionTypes,
		),
	)

	return c.AnswerCallback()
}

func mentionTypesKeyboard(
	t *i18n.Translator,
	lang string,
	types chat.MentionTypes,
) *botapi.InlineKeyboardMarkup {
	return botapi.InlineKeyboard(
		botapi.InlineRow(
			botapi.InlineButtonData(
				check(types.Has(chat.MentionEmoji))+
					" "+t.T(lang, i18n.Cmd.Summon.Style.Emoji),
				"summon:style:emoji",
			),
		),
		botapi.InlineRow(
			botapi.InlineButtonData(
				check(types.Has(chat.MentionName))+
					" "+t.T(lang, i18n.Cmd.Summon.Style.Name),
				"summon:style:name",
			),
		),
		botapi.InlineRow(
			botapi.InlineButtonData(
				check(types.Has(chat.MentionRole))+
					" "+t.T(lang, i18n.Cmd.Summon.Style.Role),
				"summon:style:role",
			),
		),
	)
}
