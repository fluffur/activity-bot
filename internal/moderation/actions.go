package moderation

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/utils"
	"activity-bot/internal/utils/tghtml"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/botapi"
)

func moderationMessage(
	loc *i18n.Localizer,
	ch chat.Chat,
	target, moderator chatmember.ChatMember,
	action i18n.MessageID,
	until time.Time,
	reason string,
	extra ...string,
) string {
	untilText := loc.T(i18n.Cmd.Moderation.Templates.Forever, nil)
	if !until.IsZero() {
		untilText = loc.T(
			i18n.Cmd.Moderation.Templates.Until,
			i18n.CmdModerationTemplatesUntilData{
				Until: tghtml.DefaultDateTime(until),
			},
		)
	}

	lines := []string{
		loc.T(
			i18n.Cmd.Moderation.Templates.Action,
			i18n.CmdModerationTemplatesActionData{
				User: tghtml.MemberMention(loc, ch, target),
				Action: loc.T(
					action,
					nil,
					i18n.WithGender(target.Gender()),
				),
				Until: untilText,
			},
		),
	}

	lines = append(lines, extra...)

	if reason != "" {
		lines = append(lines, loc.T(
			i18n.Cmd.Moderation.Templates.Reason,
			i18n.CmdModerationTemplatesReasonData{
				Reason: reason,
			},
		))
	}

	lines = append(lines, loc.T(
		i18n.Cmd.Moderation.Templates.Moderator,
		i18n.CmdModerationTemplatesModeratorData{
			Moderator: tghtml.MemberMention(loc, ch, moderator),
		},
	))

	return strings.Join(lines, "\n")
}

func (h *Handler) Ban(c *botapi.Context) error {
	moderator := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)

	target, ok := args.User()
	if !ok || moderator.ID() == target.ID() {
		return nil
	}

	until, ok := args.Until()
	if !ok {
		until = time.Time{}
	}

	reason, _ := args.Text()
	ch := cctx.MustChat(c)

	if err := h.service.Ban(c, ch.ID, target, moderator, until, reason); err != nil {
		return fmt.Errorf("service ban: %w", err)
	}

	if err := c.Bot.BanChatMember(
		c,
		botapi.ID(ch.ID),
		target.ID(),
		botapi.WithBanUntil(int(until.Unix())),
	); err != nil {
		return fmt.Errorf("bot ban: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	text := moderationMessage(
		loc,
		ch,
		target,
		moderator,
		i18n.Cmd.Moderation.Actions.Ban,
		until,
		reason,
	)

	_, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) Kick(c *botapi.Context) error {
	moderator := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)

	target, ok := args.User()
	if !ok || moderator.ID() == target.ID() {
		return nil
	}

	reason, _ := args.Text()
	ch := cctx.MustChat(c)

	if err := h.service.Kick(c, ch.ID, target, moderator, reason); err != nil {
		return fmt.Errorf("service kick: %w", err)
	}

	if err := c.Bot.BanChatMember(c, botapi.ID(ch.ID), target.ID()); err != nil {
		return fmt.Errorf("bot kick ban: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	text := moderationMessage(
		loc,
		ch,
		target,
		moderator,
		i18n.Cmd.Moderation.Actions.Kick,
		time.Time{},
		reason,
	)

	_, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) Mute(c *botapi.Context) error {
	moderator := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)

	target, ok := args.User()
	if !ok || moderator.ID() == target.ID() {
		return nil
	}

	until, ok := args.Until()
	if !ok {
		until = time.Time{}
	}

	reason, _ := args.Text()
	ch := cctx.MustChat(c)

	if err := h.service.Mute(c, ch.ID, target, moderator, until, reason); err != nil {
		return fmt.Errorf("service mute: %w", err)
	}

	permissions := botapi.ChatPermissions{
		CanSendMessages: false,
	}

	if err := c.Bot.RestrictChatMember(
		c,
		botapi.ID(ch.ID),
		target.ID(),
		permissions,
		int(until.Unix()),
	); err != nil {
		return fmt.Errorf("bot mute: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	text := moderationMessage(
		loc,
		ch,
		target,
		moderator,
		i18n.Cmd.Moderation.Actions.Mute,
		until,
		reason,
	)

	_, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) Unban(c *botapi.Context) error {
	args := cctx.MustArgs(c)

	target, ok := args.User()
	if !ok {
		return nil
	}

	ch := cctx.MustChat(c)

	if err := c.Bot.UnbanChatMember(
		c,
		botapi.ID(ch.ID),
		target.ID(),
	); err != nil {
		return fmt.Errorf("bot unmute: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	text := loc.T(i18n.Cmd.Moderation.Unban.Unbanned, i18n.CmdModerationUnbanUnbannedData{
		User: tghtml.MemberMention(loc, ch, target),
	})

	_, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) ListAdmins(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	admins, err := h.service.GetAdmins(c, ch.ID)
	if err != nil {
		return fmt.Errorf("list admins: %w", err)
	}

	if len(admins) == 0 {
		_, err = c.Reply(loc.T(i18n.Cmd.Moderation.ListAdmins.Empty, nil))
		return err
	}

	categories := map[permission.Status][]chatmember.ChatMember{}
	for _, admin := range admins {
		categories[admin.Status] = append(categories[admin.Status], admin)
	}

	order := []permission.Status{
		permission.StatusOwner,
		permission.StatusCoOwner,
		permission.StatusSeniorAdmin,
		permission.StatusAdmin,
		permission.StatusModerator,
	}

	var text strings.Builder

	text.WriteString(loc.T(i18n.Cmd.Moderation.ListAdmins.Header, nil))

	for _, status := range order {
		members := categories[status]
		if len(members) == 0 {
			continue
		}

		text.WriteString("\n\n")
		text.WriteString(
			status.Emoji() + " " + utils.UcFirst(loc.T(status.TranslationKey(), nil, i18n.WithPluralCount(len(members)))),
		)

		for i, member := range members {
			text.WriteString(fmt.Sprintf("\n%d. ", i+1))
			text.WriteString(tghtml.MemberLink(loc, ch, member))
		}
	}

	_, err = c.Reply(
		text.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func (h *Handler) SetStatus(c *botapi.Context) error {
	moderator := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)

	target, ok := args.User()
	if !ok || moderator.ID() == target.ID() {
		return nil
	}

	statusValue, ok := args.Number()
	if !ok {
		statusValue = int64(permission.StatusAdmin)
	}

	status := permission.Status(statusValue)
	if !status.IsValid() {
		return nil
	}

	if err := h.service.SetStatus(c, ch.ID, moderator, target, status); err != nil {
		switch {
		case errors.Is(err, ErrUserCantBeModerated):
			_, _ = c.Reply(loc.T(i18n.Cmd.Moderation.SetStatus.CantModerate, nil))
			return nil

		case errors.Is(err, ErrUserStatusInvalid):
			_, _ = c.Reply(loc.T(i18n.Cmd.Moderation.SetStatus.InvalidStatus, nil))
			return nil

		default:
			return fmt.Errorf("set status: %w", err)
		}
	}

	text := loc.T(i18n.Cmd.Moderation.SetStatus.Success,
		i18n.CmdModerationSetStatusSuccessData{
			User:   tghtml.MemberMention(loc, ch, target),
			Status: tghtml.Bold(tghtml.Status(loc, status)),
		},
	)

	_, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) Promote(c *botapi.Context) error {
	return h.ChangeStatus(c, +1)
}

func (h *Handler) Demote(c *botapi.Context) error {
	return h.ChangeStatus(c, -1)
}

func (h *Handler) ChangeStatus(c *botapi.Context, delta int) error {
	moderator := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)

	target, ok := args.User()
	if !ok || moderator.ID() == target.ID() {
		return nil
	}

	currentStatus := target.Status

	var newStatus permission.Status

	if value, ok := args.Number(); ok {
		newStatus = permission.Status(value)
		if !newStatus.IsValid() {
			return nil
		}
	} else {
		newStatus = currentStatus + permission.Status(delta)

		if newStatus < permission.StatusMin {
			newStatus = permission.StatusMin
		}

		if newStatus > permission.StatusMax {
			newStatus = permission.StatusMax
		}
	}

	if err := h.service.SetStatus(c, ch.ID, moderator, target, newStatus); err != nil {
		switch {
		case errors.Is(err, ErrUserCantBeModerated):
			_, _ = c.Reply(loc.T(i18n.Cmd.Moderation.SetStatus.CantModerate, nil))
			return nil

		case errors.Is(err, ErrUserStatusInvalid):
			_, _ = c.Reply(loc.T(i18n.Cmd.Moderation.SetStatus.InvalidStatus, nil))
			return nil

		default:
			return fmt.Errorf("change status: %w", err)
		}
	}

	text := loc.T(i18n.Cmd.Moderation.SetStatus.Success,
		i18n.CmdModerationSetStatusSuccessData{
			User:   tghtml.MemberMention(loc, ch, target),
			Status: tghtml.Bold(tghtml.Status(loc, newStatus)),
		},
	)

	_, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}

func (h *Handler) RemoveAdmin(c *botapi.Context) error {
	moderator := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)
	loc := cctx.MustLocalizer(c)
	ch := cctx.MustChat(c)

	target, ok := args.User()
	if !ok || moderator.ID() == target.ID() {
		return nil
	}

	if err := h.service.SetStatus(
		c,
		ch.ID,
		moderator,
		target,
		permission.StatusMember,
	); err != nil {
		switch {
		case errors.Is(err, ErrUserCantBeModerated):
			_, _ = c.Reply(loc.T(i18n.Cmd.Moderation.RemoveAdmin.CantModerate, nil))
			return nil

		default:
			return fmt.Errorf("remove admin: %w", err)
		}
	}

	text := loc.T(
		i18n.Cmd.Moderation.RemoveAdmin.Success,
		i18n.CmdModerationRemoveAdminSuccessData{
			User: tghtml.MemberMention(loc, ch, target),
		},
	)

	_, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML))

	return err
}
