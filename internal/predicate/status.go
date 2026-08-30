package predicate

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"activity-bot/internal/utils/tghtml"
	"context"
	"errors"

	glog "github.com/gotd/log"
	"github.com/jackc/pgx/v5"

	"github.com/gotd/botapi"
)

type PermissionChecker struct {
	repo  permission.Repository
	devID int64
}

func NewPermissionsChecker(
	repo permission.Repository,
	devID int64,
) *PermissionChecker {
	return &PermissionChecker{
		repo:  repo,
		devID: devID,
	}
}

func (p *PermissionChecker) Pass(
	name string,
	defaultStatus permission.Status,
) botapi.Predicate {
	return func(c *botapi.Context) bool {
		status, ok := p.status(c, name, defaultStatus)
		if !ok {
			return false
		}

		c.Context = cctx.WithCommandPermission(c.Context, status)

		return true
	}
}

func (p *PermissionChecker) Require(
	name string,
	defaultStatus permission.Status,
) botapi.Predicate {
	return func(c *botapi.Context) bool {
		ch, err := cctx.Chat(c)
		if err != nil {
			glog.For(c.Bot.Logger()).Error(
				c,
				"status no chat",
				glog.Error(err),
			)
		}
		if ch.ID == 0 {
			return true
		}

		status, ok := p.status(c, name, defaultStatus)
		if !ok {
			return false
		}

		cctx.WithCommandPermission(c, status)

		cm, err := cctx.ChatMember(c)
		if err != nil {
			glog.For(c.Bot.Logger()).Error(
				c,
				"status no chat member",
				glog.Error(err),
			)

			return false
		}

		loc := cctx.MustLocalizer(c)

		if cctx.DevPassed(c) {
			return true
		}

		if cm.Permitted(status) {
			return true
		}

		if c.Update.CallbackQuery != nil {
			_ = c.AnswerCallback(
				botapi.WithCallbackText(
					loc.T(
						i18n.System.NoPermission,
						i18n.SystemNoPermissionData{
							Status: loc.T(status.TranslationKey(), nil),
						},
					),
				),
			)

			return false
		}

		if c.Update.Message != nil {
			if status.IsDisabled() {
				return false
			}

			_, _ = c.Reply(
				loc.T(
					i18n.System.NoPermission,
					i18n.SystemNoPermissionData{
						Status: tghtml.Bold(
							loc.T(status.TranslationKey(), nil),
						),
					},
				),
				botapi.WithParseMode(botapi.ParseModeHTML),
			)

			return false
		}

		return false
	}
}

func getStatus(
	ctx context.Context,
	chatID int64,
	name string,
	defaultStatus permission.Status,
	repo permission.Repository,
) (permission.Status, error) {
	status, err := repo.CommandPermission(ctx, chatID, name)
	if err == nil {
		return status, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return defaultStatus, nil
	}

	return 0, err
}

func (p *PermissionChecker) status(
	c *botapi.Context,
	name string,
	defaultStatus permission.Status,
) (permission.Status, bool) {
	ch, err := cctx.Chat(c)
	if err != nil {
		glog.For(c.Bot.Logger()).Error(c, "status no chat", glog.Error(err))

		return 0, false
	}

	if ch.ID == 0 {
		return defaultStatus, true
	}

	if p.repo == nil {
		return defaultStatus, true
	}

	status, err := getStatus(c, ch.ID, name, defaultStatus, p.repo)
	if err != nil {
		glog.For(c.Bot.Logger()).Error(c, "get status error", glog.Error(err))

		return 0, false
	}

	return status, true
}

func (p *PermissionChecker) PassDev() botapi.Predicate {
	return func(c *botapi.Context) bool {
		msg := c.Message()
		if msg == nil {
			return false
		}
		var senderID int64
		sender := c.Sender()
		if sender != nil {
			senderID = sender.ID
		} else {
			senderID = msg.Chat.ID
		}

		if p.devID == senderID {
			c.Context = cctx.PassDev(c.Context)
		}

		return true
	}
}

func (p *PermissionChecker) RequireDev() botapi.Predicate {
	return func(c *botapi.Context) bool {
		msg := c.Message()
		if msg == nil {
			return false
		}

		var senderID int64

		sender := c.Sender()
		if sender != nil {
			senderID = sender.ID
		} else {
			senderID = msg.Chat.ID
		}

		if p.devID != senderID {
			return false
		}

		c.Context = cctx.PassDev(c.Context)

		return true
	}
}
