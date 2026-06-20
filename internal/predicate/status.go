package predicate

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
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
			chatID, ok := c.Chat()
			if !ok {
				return false
			}

			text := p.translator.TData(ch.Language, i18n.NoPermission, i18n.NoPermissionArgs(
				p.translator.T(ch.Language, status.TranslationKey())),
			)

			if c.Update.CallbackQuery != nil {
				_ = c.Bot.AnswerCallbackQuery(c.Context,
					c.Update.CallbackQuery.ID, botapi.WithCallbackText(text))
			} else if c.Update.Message != nil {
				_, _ = c.Bot.SendMessage(c.Context, chatID, text, botapi.ReplyTo(c.Message().MessageID))
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
