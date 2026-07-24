package moderation

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gotd/td/tgerr"

	"github.com/gotd/botapi"
)

func isValidRoleString(text string) bool {
	return utf8.RuneCountInString(text) <= 16
}

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

	chatID, _ := c.Chat()

	if err := h.roleUpdater.Update(c, int64(chatID.(botapi.ChatIDInt)), c.Bot); err != nil {
		return fmt.Errorf("set role: %w", err)
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

func (h *Handler) ListRoles(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	members, err := h.chatMemberService.ListHumanPresentChatMembers(c, ch.ID)
	if err != nil {
		return fmt.Errorf("list roles members: %w", err)
	}

	var withRoles []chatmember.ChatMember
	for _, member := range members {
		if member.Tag != "" {
			withRoles = append(withRoles, member)
		}
	}

	if len(withRoles) == 0 {
		_, err := c.Reply(loc.T(i18n.Cmd.Moderation.ListRoles.Empty, nil))
		return err
	}

	slices.SortFunc(withRoles, func(a, b chatmember.ChatMember) int {
		return strings.Compare(a.Tag, b.Tag)
	})

	var text strings.Builder

	text.WriteString(loc.T(i18n.Cmd.Moderation.ListRoles.Header, nil))
	text.WriteString("\n\n")
	text.WriteString("<blockquote expandable>")

	for i, member := range withRoles {
		if i > 0 {
			text.WriteByte('\n')
		}

		text.WriteString(fmt.Sprintf("%d. ", i+1))
		text.WriteString(tghtml.MemberLink(loc, ch, member))
		if member.User.Username != "" {
			text.WriteString(fmt.Sprintf(" <code>@%s</code>", member.User.Username))
		}
	}
	text.WriteString("</blockquote>")

	_, err = c.Reply(
		text.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}
