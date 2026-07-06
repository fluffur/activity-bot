package moderation

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/gotd/td/tgerr"

	"github.com/gotd/botapi"
)

func (h *Handler) SetRole(c *botapi.Context) error {
	cm := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)
	ch := cctx.MustChat(c)

	u, ok := args.User()
	if ok {
		cm = u
	}

	newRole, ok := args.Text()
	if !ok {
		return nil
	}

	if err := h.chatMemberService.UpdateTag(c, ch.ID, cm.ID(), newRole); err != nil {
		return fmt.Errorf("set role upd tag: %w", err)
	}

	loc := cctx.MustLocalizer(c)
	if err := c.Bot.SetChatMemberTag(c, botapi.ID(ch.ID), cm.ID(), newRole); err != nil {
		if tgErr, ok := errors.AsType[*tgerr.Error](err); ok && tgErr.Type == "CHAT_CREATOR_REQUIRED" {
			if cm.IsOwner() {
				_, err = c.Reply(loc.T(i18n.Cmd.Moderation.SetRole.ChatCreator, nil))
			} else {
				_, err = c.Reply(loc.T(i18n.Cmd.Moderation.SetRole.ChatAdmin, nil))
			}

			return err
		}

		if tgErr, ok := errors.AsType[*tgerr.Error](err); ok && tgErr.Type == "CHAT_ADMIN_REQUIRED" {
			_, err = c.Reply(loc.T(i18n.Cmd.Moderation.SetRole.NoRights, nil))

			return err
		}

		return fmt.Errorf("set role set tag: %w", err)
	}

	if _, err := c.Reply(
		loc.T(i18n.Cmd.Moderation.SetRole.Set, i18n.CmdModerationSetRoleAdminSetData{
			User: tghtml.MemberMention(loc, ch, cm),
		}),
		botapi.WithParseMode(botapi.ParseModeHTML),
	); err != nil {
		return fmt.Errorf("set role reply: %w", err)
	}

	return nil
}

func (h *Handler) SetRoleAdmin(c *botapi.Context) error {
	cm := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)
	ch := cctx.MustChat(c)

	u, ok := args.User()
	if ok {
		cm = u
	}

	newRole, ok := args.Text()
	if !ok {
		return nil
	}

	if err := h.chatMemberService.UpdateTag(c, ch.ID, cm.ID(), newRole); err != nil {
		return fmt.Errorf("set role admin upd tag: %w", err)
	}

	loc := cctx.MustLocalizer(c)
	if err := c.Bot.SetChatMemberTag(c, botapi.ID(ch.ID), cm.ID(), newRole); err != nil {
		if tgErr, ok := errors.AsType[*tgerr.Error](err); ok && tgErr.Type == "CHAT_CREATOR_REQUIRED" {
			if cm.IsOwner() {
				_, err = c.Reply(loc.T(i18n.Cmd.Moderation.SetRoleAdmin.ChatCreator, nil))
			} else {
				_, err = c.Reply(loc.T(i18n.Cmd.Moderation.SetRoleAdmin.ChatAdmin, nil))
			}

			return err
		}

		if tgErr, ok := errors.AsType[*tgerr.Error](err); ok && tgErr.Type == "CHAT_ADMIN_REQUIRED" {
			_, err = c.Reply(loc.T(i18n.Cmd.Moderation.SetRoleAdmin.NoRights, nil))

			return err
		}

		return fmt.Errorf("set role admin set tag: %w", err)
	}

	if err := c.Bot.PromoteChatMember(c, botapi.ID(ch.ID), cm.ID(), botapi.ChatAdminRights{
		CanManageChat: true,
		CustomTitle:   newRole,
	}); err != nil {
		return fmt.Errorf("set role admin promote chat member: %w", err)
	}

	if _, err := c.Reply(
		loc.T(i18n.Cmd.Moderation.SetRoleAdmin.Set, i18n.CmdModerationSetRoleAdminSetData{
			User: tghtml.MemberMention(loc, ch, cm),
		}),
		botapi.WithParseMode(botapi.ParseModeHTML),
	); err != nil {
		return fmt.Errorf("set role admin reply: %w", err)
	}

	return nil
}

func isValidRoleString(text string) bool {
	return utf8.RuneCountInString(text) <= 16
}
