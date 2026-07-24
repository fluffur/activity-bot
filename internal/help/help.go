package help

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"

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
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	eb := &entity.Builder{}
	eb.CustomEmoji("👋", 5260536644913604662)
	message, entities := eb.Complete()
	peer, err := resolveChannel(c, c.Bot.Raw(), "asadsaas")
	if err != nil {
		return err
	}

	_, err = c.Bot.Raw().MessagesSendMessage(c, &tg.MessagesSendMessageRequest{
		Message:  message,
		Peer:     peer,
		Entities: entities,
		RandomID: rand.Int63(),
	})
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	_, err = c.Bot.SendMessage(c, botapi.Username("asadsaas"), "<tg-emoji emoji-id=\"5260536644913604662\">👋</tg-emoji>",
		botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func resolveChannel(ctx context.Context, api *tg.Client, username string) (tg.InputPeerClass, error) {
	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}

	switch peer := resolved.Peer.(type) {
	case *tg.PeerChannel:
		for _, ch := range resolved.Chats {
			channel, ok := ch.(*tg.Channel)
			if !ok {
				continue
			}

			if channel.ID == peer.ChannelID {
				return &tg.InputPeerChannel{
					ChannelID:  channel.ID,
					AccessHash: channel.AccessHash,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("channel not found")
}

func (h *Handler) ShowCategory(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)

	parts := strings.Split(
		c.Update.CallbackQuery.Data,
		":",
	)

	if len(parts) < 2 {
		return c.AnswerCallback()
	}

	category := command.Category(parts[2])

	page := 0
	if len(parts) >= 4 {
		p, err := strconv.Atoi(parts[3])
		if err == nil {
			page = p
		}
	}

	chatID, _ := c.Chat()
	ch := cctx.MustChat(c)

	permissions, err := h.permissionRepo.CommandPermissions(c, ch.ID)
	if err != nil {
		return err
	}
	_, err = c.Bot.EditMessageText(
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		h.renderCategory(loc, category, page, permissions),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.commandsKeyboard(loc, category, page),
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

	ch := cctx.MustChat(c)

	permissions, err := h.permissionRepo.CommandPermissions(c, ch.ID)
	if err != nil {
		return err
	}
	_, err = c.Bot.EditMessageText(
		c,
		chatID,
		c.Update.CallbackQuery.Message.MessageID,
		h.renderCommand(loc, cmd, permissions),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.commandKeyboard(loc, category, key),
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

	name, ok := args.Text()
	if !ok {
		return nil
	}

	name = strings.ToLower(name)

	cmd, ok := h.registry.FindByKeyOrAlias(name)

	if !ok {
		return nil
	}

	ch := cctx.MustChat(c)

	permissions, err := h.permissionRepo.CommandPermissions(c, ch.ID)
	if err != nil {
		return err
	}

	_, err = c.Reply(
		h.renderCommand(loc, cmd, permissions),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(
			h.commandKeyboard(loc, cmd.Category, cmd.Key),
		),
	)

	return err
}
