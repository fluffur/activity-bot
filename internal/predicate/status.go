package predicate

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"context"
	"errors"

	glog "github.com/gotd/log"
	"github.com/jackc/pgx/v5"

	"github.com/gotd/botapi"
)

type PermissionRepository interface {
	CommandPermission(ctx context.Context, chatID int64, name string) (chatmember.Status, error)
}

type PermissionChecker struct {
	repo       PermissionRepository
	translator *i18n.Translator
}

func NewPermissionsChecker(repo PermissionRepository, translator *i18n.Translator) *PermissionChecker {
	return &PermissionChecker{repo, translator}
}

func (p *PermissionChecker) Pass(
	name string,
	defaultStatus chatmember.Status,
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
	defaultStatus chatmember.Status,
) botapi.Predicate {
	return func(c *botapi.Context) bool {
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
	defaultStatus chatmember.Status,
	repo PermissionRepository,
) (chatmember.Status, error) {
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
	defaultStatus chatmember.Status,
) (chatmember.Status, bool) {
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
