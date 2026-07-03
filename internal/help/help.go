package help

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"github.com/gotd/botapi"
)

func (h *Handler) Help(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(
		h.renderCategories(loc),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.categoriesKeyboard(loc),
		),
	)

	return err
}

func (h *Handler) ShowCategory(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)

	data := c.Update.CallbackQuery.Data

	category := command.Category(
		strings.TrimPrefix(
			data,
			callbackHelpCategory+":",
		),
	)

	chatID, _ := c.Chat()

	_, err := c.Bot.EditMessageText(
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		h.renderCategory(loc, category),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.commandsKeyboard(loc, category),
		),
	)

	if err != nil {
		return fmt.Errorf("edit help category: %w", err)
	}

	return c.AnswerCallback()
}

func (h *Handler) ShowCategories(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)

	chatID, _ := c.Chat()

	_, err := c.Bot.EditMessageText(
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		h.renderCategories(loc),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.categoriesKeyboard(loc),
		),
	)

	if err != nil {
		return fmt.Errorf("edit help categories: %w", err)
	}

	return c.AnswerCallback()
}

func (h *Handler) ShowCommand(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)

	data := strings.Split(
		c.Update.CallbackQuery.Data,
		":",
	)

	spew.Dump(data)
	if len(data) != 4 {
		return c.AnswerCallback()
	}

	category := command.Category(data[2])
	key := data[3]

	cmd, ok := h.registry.Find(key)
	if !ok {
		return c.AnswerCallback()
	}

	chatID, _ := c.Chat()
	spew.Dump(h.renderCommand(loc, cmd))
	_, err := c.Bot.EditMessageText(
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		h.renderCommand(loc, cmd),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.commandKeyboard(category, key),
		),
	)

	if err != nil {
		return fmt.Errorf("edit help command: %w", err)
	}

	return c.AnswerCallback()
}

func (h *Handler) ShowCommandHelp(
	c *botapi.Context,
) error {
	loc := cctx.MustLocalizer(c)
	args := cctx.MustArgs(c)

	if len(args.Texts[0]) == 0 {
		return nil
	}

	name := strings.ToLower(args.Texts[0])
	cmd, ok := h.registry.FindByKeyOrAlias(name)
	if !ok {
		return nil
	}

	_, err := c.Reply(
		h.renderCommand(loc, cmd),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.commandKeyboard(cmd.Category, cmd.Key),
		),
	)

	return err
}
