package handler

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/utils/tghtml"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) ShowPermission(c *botapi.Context) error {
	ch := cctx.MustChat(c)

	text, ok := cctx.MustArgs(c).Text()
	if !ok {
		return nil
	}

	cmd, ok := h.registry.FindByKeyOrAlias(text)
	if !ok {
		return nil
	}

	status, err := h.repo.CommandPermission(c, ch.ID, cmd.Key)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("get permission: %w", err)
		}

		status = cmd.Permission
	}

	loc := cctx.MustLocalizer(c)
	statusStr := loc.T(status.TranslationKey(), nil)

	_, err = c.Reply(loc.T(i18n.Cmd.Permission.Show.Success, i18n.CmdPermissionShowSuccessData{
		Cmd:    tghtml.Bold(text + " " + "(" + cmd.Key + ")"),
		Status: status.Emoji() + " " + tghtml.Bold(statusStr),
	}),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func (h *Handler) SetPermission(c *botapi.Context) error {
	ch := cctx.MustChat(c)

	text, ok := cctx.MustArgs(c).Text()
	if !ok {
		return nil
	}

	number, ok := cctx.MustArgs(c).Number()
	if !ok {
		return nil
	}

	status := permission.Status(number)
	if !status.IsValid() {
		status = permission.StatusDisabled
	}

	cmd, ok := h.registry.FindByKeyOrAlias(text)
	if !ok {
		return fmt.Errorf("no such permission")
	}
	loc := cctx.MustLocalizer(c)

	if cmd.Key == "set_permission" {
		switch status {
		case permission.StatusMember:
			_, err := c.Reply(loc.T(i18n.Cmd.Permission.Set.CannotAllowMembers, nil))
			return err
		case permission.StatusDisabled:
			_, err := c.Reply(loc.T(i18n.Cmd.Permission.Set.CannotDisable, nil))
			return err
		default:
		}
	}

	if err := h.repo.SetCommandPermission(c, ch.ID, cmd.Key, status); err != nil {
		return fmt.Errorf("set permission: %w", err)
	}

	statusStr := loc.T(status.TranslationKey(), nil)

	_, err := c.Reply(loc.T(i18n.Cmd.Permission.Set.Success, i18n.CmdPermissionShowSuccessData{
		Cmd:    tghtml.Bold(text + " " + "(" + cmd.Key + ")"),
		Status: status.Emoji() + " " + tghtml.Bold(statusStr),
	}),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}
