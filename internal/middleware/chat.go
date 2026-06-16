package middleware

import (
	db "activity-bot/internal/db/postgres/sqlc"

	"github.com/gotd/botapi"
)

func Chat(queries *db.Queries) botapi.Middleware {
	return func(next botapi.Handler) botapi.Handler {
		return func(c *botapi.Context) error {

			err := next(c)
			return err
		}
	}
}
