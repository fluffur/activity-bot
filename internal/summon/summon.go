package summon

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

func (h *Handler) SummonAll(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("summon chat: %w", err)
	}

	args := predicate.Args(c)

	var text string
	if args != nil {
		text = ReplaceMentions(args.OriginalTextHTML(), args.Entities)
	}

	if strings.TrimSpace(text) == "" {
		text = ch.WelcomeCallMessage
	}

	cms, err := h.chatMemberService.ListSummonChatMembers(c.Context, ch.ID)
	if err != nil {
		return fmt.Errorf("summon cms list: %w", err)
	}

	return h.Summon(c, text, ch, cms)
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
