package moderation

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/utils/tghtml"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/botapi"
)

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

func (h *Handler) Warn(c *botapi.Context) error {
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

	warnsCount, err := h.service.Warn(
		c,
		ch,
		target,
		moderator,
		reason,
		until,
	)
	if err != nil {
		return fmt.Errorf("service warn: %w", err)
	}

	if warnsCount >= ch.MaxWarns {
		if err := h.service.Ban(c, ch.ID, target, moderator, until, reason); err != nil {
			return fmt.Errorf("service auto ban: %w", err)
		}

		if err := c.Bot.BanChatMember(
			c,
			botapi.ID(ch.ID),
			target.ID(),
			botapi.WithBanUntil(int(until.Unix())),
		); err != nil {
			return fmt.Errorf("bot auto ban: %w", err)
		}

		loc := cctx.MustLocalizer(c)

		_, err := c.Reply(
			moderationMessage(
				loc,
				ch,
				target,
				moderator,
				i18n.Cmd.Moderation.Actions.Ban,
				until,
				reason,
			),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	loc := cctx.MustLocalizer(c)

	_, err = c.Reply(
		moderationMessage(
			loc,
			ch,
			target,
			moderator,
			i18n.Cmd.Moderation.Actions.Warn,
			until,
			reason,
			loc.T(
				i18n.Cmd.Moderation.Templates.Warns,
				i18n.CmdModerationTemplatesWarnsData{
					Current: warnsCount,
					Max:     ch.MaxWarns,
				},
			),
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
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
		return nil
	}

	if !permission.IsValidStatus(statusValue) {
		return nil
	}

	status := permission.Status(statusValue)

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

func moderationMessage(
	loc *i18n.Localizer,
	ch chat.Chat,
	target, moderator chatmember.ChatMember,
	action i18n.GenderMessage,
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
				Action: loc.TGender(
					target.Gender(),
					action,
					nil,
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
