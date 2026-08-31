package middleware

import (
	"activity-bot/internal/ban"
	"activity-bot/internal/cctx"
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/gotd/botapi"
	"github.com/jackc/pgx/v5"
)

func BanMiddleware(repo ban.Repository) botapi.Middleware {
	return func(next botapi.Handler) botapi.Handler {
		return func(c *botapi.Context) error {
			chat, err := cctx.Chat(c)
			if err != nil {
				return err
			}

			if sender := c.Sender(); sender != nil {
				userBan, err := repo.GetUserBan(c, sender.ID)
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("get user ban: %w", err)
				}

				if err == nil {
					spew.Dump("user banned", userBan, c.Message().Text)

					_, err := c.Reply(
						fmt.Sprintf(
							"Вы забанены у бота.%s",
							formatBanReason(userBan.Reason),
						),
					)

					return err
				}
			}

			chatBan, err := repo.GetChatBan(c, chat.ID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("get chat ban: %w", err)
			}

			if err == nil {
				_, err := c.Reply(
					fmt.Sprintf(
						"Чат забанен.%s",
						formatBanReason(chatBan.Reason),
					),
				)

				return err
			}

			return next(c)
		}
	}
}

func formatBanReason(reason string) string {
	if reason == "" {
		return ""
	}

	return fmt.Sprintf("\nПричина: %s", reason)
}
