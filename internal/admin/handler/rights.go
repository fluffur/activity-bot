package handler

import (
	"activity-bot/internal/admin/view"
	"activity-bot/internal/command"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"strconv"
	"strings"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

func (h *Handler) ManageRights(ctx *command.Context, u *ext.Update) error {
	eb := &entity.Builder{}
	view.WriteCategories(eb)
	return ctx.ReplyOnly(u, options.WithBuilder(eb), options.WithMarkup(view.GetCategoriesKeyboard()))
}

func (h *Handler) CallbackManageRights(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	data := u.CallbackQuery.Data
	parts := strings.Split(string(data), ":")
	action := parts[0]

	switch action {
	case "rights_list":
		eb := &entity.Builder{}
		view.WriteCategories(eb)
		text, entities := eb.Complete()
		_, err := ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
			ID:          u.CallbackQuery.GetMsgID(),
			Message:     text,
			Entities:    entities,
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
		eb := &entity.Builder{}
		view.WriteCommandsByCategory(eb, category, configCmds, perms)
		text, entities := eb.Complete()
		_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
			ID:          u.CallbackQuery.GetMsgID(),
			Message:     text,
			Entities:    entities,
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
		eb := &entity.Builder{}
		view.WriteEditCommandRights(eb, key, status, configCmds)
		text, entities := eb.Complete()

		_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
			ID:          u.CallbackQuery.GetMsgID(),
			Message:     text,
			Entities:    entities,
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

		configCmds := h.factory.ConfigurableCommands()
		eb := &entity.Builder{}
		view.WriteEditCommandRights(eb, key, status, configCmds)
		text, entities := eb.Complete()

		_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
			ID:          u.CallbackQuery.GetMsgID(),
			Message:     text,
			Entities:    entities,
			ReplyMarkup: view.GetEditRightsKeyboard(key, status, configCmds),
		})
		if err == nil {
			_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				Message: "Права обновлены",
			})
		} else {
			_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
				Message: "Права с этим уровнем уже установлены",
			})
		}
		return err
	}

	return nil
}
