package handler

import (
	"activity-bot/internal/admin/view"
	"activity-bot/internal/command"
	"activity-bot/internal/model"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func (h *Handler) ManageRights(b *gotgbot.Bot, ctx *command.Context) error {
	return ctx.Reply(b, view.FormatCategories(), &gotgbot.SendMessageOpts{
		ReplyMarkup: view.GetCategoriesKeyboard(),
		ParseMode:   gotgbot.ParseModeHTML,
	})
}

func (h *Handler) CallbackManageRights(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	data := ctx.CallbackQuery.Data
	parts := strings.Split(data, ":")
	action := parts[0]

	switch action {
	case "rights_list":
		_, _, err := ctx.CallbackQuery.Message.EditText(b, view.FormatCategories(), &gotgbot.EditMessageTextOpts{
			ParseMode:   gotgbot.ParseModeHTML,
			ReplyMarkup: view.GetCategoriesKeyboard(),
		})
		return err

	case "rights_cat":
		if len(parts) < 2 {
			return nil
		}
		category := command.Category(parts[1])
		perms, err := h.chatService.GetCommandPermissions(ctx.StdContext(), c.ID)
		if err != nil {
			return err
		}

		configCmds := h.factory.ConfigurableCommands()
		_, _, err = ctx.CallbackQuery.Message.EditText(b, view.FormatCommandsByCategory(category, configCmds, perms), &gotgbot.EditMessageTextOpts{
			ParseMode:   gotgbot.ParseModeHTML,
			ReplyMarkup: view.GetCommandsByCategoryKeyboard(category, configCmds, perms),
		})
		return err

	case "rights_edit":
		if len(parts) < 2 {
			return nil
		}
		key := parts[1]
		status, err := h.chatService.GetCommandPermission(ctx.StdContext(), c.ID, key)
		configCmds := h.factory.ConfigurableCommands()
		if err != nil {
			status = command.GetDefaultStatus(configCmds, key)
		}

		_, _, err = ctx.CallbackQuery.Message.EditText(b, view.FormatEditCommandRights(key, status, configCmds), &gotgbot.EditMessageTextOpts{
			ParseMode:   gotgbot.ParseModeHTML,
			ReplyMarkup: view.GetEditRightsKeyboard(key, status, configCmds),
		})
		return err

	case "rights_set":
		if len(parts) < 3 {
			return nil
		}
		key := parts[1]
		statusInt, _ := strconv.Atoi(parts[2])
		status := model.Status(statusInt)

		err := h.chatService.SetCommandPermission(ctx.StdContext(), c.ID, key, status)
		if err != nil {
			return err
		}

		var category command.Category
		configCmds := h.factory.ConfigurableCommands()
		for _, c := range configCmds {
			if c.Name() == key {
				category = c.Category()
				break
			}
		}

		perms, err := h.chatService.GetCommandPermissions(ctx.StdContext(), c.ID)
		if err != nil {
			return err
		}
		_, _, err = ctx.CallbackQuery.Message.EditText(b, view.FormatCommandsByCategory(category, configCmds, perms), &gotgbot.EditMessageTextOpts{
			ParseMode:   gotgbot.ParseModeHTML,
			ReplyMarkup: view.GetCommandsByCategoryKeyboard(category, configCmds, perms),
		})
		if err == nil {
			_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Права обновлены"})
		}
		return err
	}

	return nil
}
