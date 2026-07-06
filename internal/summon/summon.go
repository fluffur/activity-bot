package summon

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/i18n"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

func (h *Handler) SummonAll(c *botapi.Context) error {
	ch := cctx.MustChat(c)

	m := c.Message()
	if m == nil {
		return nil
	}

	args, _ := cctx.ArgsMessage(c)

	if !ch.SkipCallConfirmation {
		return h.RequestSummonConfirmation(c, TextFromArgs(ch, args), m.MessageID)
	}

	cms, err := h.chatMemberService.ListSummonChatMembers(c, ch.ID)
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
	ch := cctx.MustChat(c)
	cm := cctx.MustChatMember(c)
	loc := cctx.MustLocalizer(c)

	if err := h.summonFSM.Enter(c, StateAwaitConfirmation, StateData{
		Text:      text,
		MessageID: msgID,
		ChatID:    ch.ID,
		UserID:    cm.ID(),
	}); err != nil {
		return err
	}

	_, err := c.Reply(
		loc.T(i18n.Cmd.Summon.Confirm.Text, nil),
		botapi.WithReplyMarkup(
			botapi.InlineKeyboard(
				botapi.InlineRow(
					botapi.InlineButtonData(
						loc.T(i18n.Cmd.Summon.Confirm.Yes, nil),
						callbackSummonConfirm,
					),
					botapi.InlineButtonData(
						loc.T(i18n.Cmd.Summon.Confirm.No, nil),
						callbackSummonCancel,
					),
				),
				botapi.InlineRow(
					botapi.InlineButtonData(
						loc.T(i18n.Cmd.Summon.Confirm.YesAndDisable, nil),
						callbackSummonConfirmDontAsk,
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
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
	)

	cms, err := h.chatMemberService.ListSummonChatMembers(c, ch.ID)
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
	loc := cctx.MustLocalizer(c)
	if err := h.summonFSM.Clear(c); err != nil {
		return err
	}

	chatID, _ := c.Chat()

	text := loc.T(i18n.Cmd.Summon.Confirm.Canceled, nil)

	_, _ = c.Bot.EditMessageText(
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		text,
	)

	return c.AnswerCallback(
		botapi.WithCallbackText(text),
	)
}

func (h *Handler) ConfirmSummonDontAsk(c *botapi.Context) error {
	data, ch, err := h.summonSession(c)
	if err != nil || data == nil {
		return err
	}

	loc := cctx.MustLocalizer(c)

	if err := h.chatService.SetSkipSummonConfirmation(c, ch.ID, true); err != nil {
		return fmt.Errorf("update chat: %w", err)
	}

	if err := h.summonFSM.Clear(c); err != nil {
		return err
	}

	text := loc.T(i18n.Cmd.Summon.Confirm.Disabled, nil)

	_ = c.AnswerCallback(
		botapi.WithCallbackText(text),
	)

	chatID, _ := c.Chat()

	_, _ = c.Bot.EditMessageText(
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		text,
	)

	cms, err := h.chatMemberService.ListSummonChatMembers(c, ch.ID)
	if err != nil {
		return fmt.Errorf("summon cms list: %w", err)
	}

	return h.Summon(c, data.Text, data.MessageID, ch, cms)
}

func (h *Handler) SummonStyle(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(
		loc.T(i18n.Cmd.Summon.Style.Text, nil),
		botapi.WithReplyMarkup(
			mentionTypesKeyboard(loc, ch.MentionTypes),
		),
	)

	return err
}

func (h *Handler) ToggleSummonStyle(c *botapi.Context) error {
	style := strings.TrimPrefix(c.Update.CallbackQuery.Data, callbackSummonStyle)

	var flag chat.MentionTypes

	switch style {
	case "emoji":
		flag = chat.MentionEmoji
	case "name":
		flag = chat.MentionName
	case "role":
		flag = chat.MentionRole
	default:
		return nil
	}

	return h.toggleMentionStyle(c, flag)
}

func (h *Handler) toggleMentionStyle(c *botapi.Context, flag chat.MentionTypes) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	if ch.MentionTypes.Has(flag) {
		ch.MentionTypes.Remove(flag)
	} else {
		ch.MentionTypes.Add(flag)
	}

	if err := h.chatService.SetMentionTypes(c, ch.ID, ch.MentionTypes); err != nil {
		return err
	}

	_, _ = c.Bot.EditMessageReplyMarkup(
		c,
		botapi.ID(ch.ID),
		c.Update.CallbackQuery.Message.MessageID,
		mentionTypesKeyboard(loc, ch.MentionTypes),
	)

	return c.AnswerCallback()
}
func mentionTypesKeyboard(
	loc *i18n.Localizer,
	types chat.MentionTypes,
) *botapi.InlineKeyboardMarkup {
	return botapi.InlineKeyboard(
		botapi.InlineRow(
			botapi.InlineButtonData(
				check(types.Has(chat.MentionEmoji))+" "+loc.T(i18n.Cmd.Summon.Style.Emoji, nil),
				"summon:style:emoji",
			),
		),
		botapi.InlineRow(
			botapi.InlineButtonData(
				check(types.Has(chat.MentionName))+" "+loc.T(i18n.Cmd.Summon.Style.Name, nil),
				"summon:style:name",
			),
		),
		botapi.InlineRow(
			botapi.InlineButtonData(
				check(types.Has(chat.MentionRole))+" "+loc.T(i18n.Cmd.Summon.Style.Role, nil),
				"summon:style:role",
			),
		),
	)
}
