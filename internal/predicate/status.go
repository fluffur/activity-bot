package predicate

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
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

func (p *PermissionChecker) Require(
	name string,
	defaultStatus chatmember.Status,
) botapi.Predicate {
	return func(c *botapi.Context) bool {
		ch, err := cctx.Chat(c.Context)
		if err != nil {
			glog.For(c.Bot.Logger()).Error(
				c.Context,
				"status no chat",
				glog.Error(err),
			)

			return false
		}
		if ch.ID == 0 {
			return true
		}

		cm, err := cctx.ChatMember(c.Context)
		if err != nil {
			glog.For(c.Bot.Logger()).Error(
				c.Context,
				"status no chat member",
				glog.Error(err),
			)

			return false
		}

		status := defaultStatus

		if p.repo != nil {
			status, err = getStatus(c.Context, ch.ID, name, defaultStatus, p.repo)

			if err != nil {
				glog.For(c.Bot.Logger()).Error(
					c.Context,
					"get status error",
					glog.Error(err),
				)

				return false
			}
		}

		if !cm.Permitted(status) {
			if c.Update.CallbackQuery != nil {
				_ = c.AnswerCallback(botapi.WithCallbackText(p.translator.TData(ch.Lang, i18n.System.NoPermission, i18n.SystemNoPermissionArgs(
					p.translator.T(ch.Lang, status.TranslationKey())),
				)))
			} else if c.Update.Message != nil {
				_, _ = c.Reply(p.translator.TData(ch.Lang, i18n.System.NoPermission, i18n.SystemNoPermissionArgs(
					tghtml.Bold(p.translator.T(ch.Lang, status.TranslationKey()))),
				), botapi.WithParseMode(botapi.ParseModeHTML))
			}
		}

		return cm.Permitted(status)
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
