package moderation

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/utils/tghtml"
	"errors"
	"fmt"
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

	if err := h.bot.BanChatMember(c, botapi.ID(ch.ID), target.ID(), botapi.WithBanUntil(int(until.Unix()))); err != nil {
		return fmt.Errorf("bot ban: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	var untilText string
	if until.IsZero() {
		untilText = loc.T(i18n.Cmd.Moderation.Ban.Forver, nil)
	} else {
		untilText = loc.T(i18n.Cmd.Moderation.Ban.Until, i18n.CmdModerationBanUntilData{Until: tghtml.DefaultDateTime(until)})
	}

	moderatorText := tghtml.MemberMention(loc, ch, moderator)
	targetText := tghtml.MemberMention(loc, ch, target)

	var text string
	if reason == "" {
		text = loc.TGender(target.Gender(), i18n.Cmd.Moderation.Ban.Banned, i18n.CmdModerationBanBannedMaleData{
			Until:     untilText,
			Moderator: moderatorText,
			User:      targetText,
		})
	} else {
		text = loc.TGender(target.Gender(), i18n.Cmd.Moderation.Ban.BannedReason, i18n.CmdModerationBanBannedReasonMaleData{
			Until:     untilText,
			Moderator: moderatorText,
			User:      targetText,
			Reason:    reason,
		})
	}

	if _, err := c.Reply(text, botapi.WithParseMode(botapi.ParseModeHTML)); err != nil {
		return fmt.Errorf("ban reply: %w", err)
	}

	return nil
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
