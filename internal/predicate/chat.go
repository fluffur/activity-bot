package predicate

import (
	"activity-bot/internal/cctx"

	"github.com/gotd/log"

	"github.com/gotd/botapi"
)

func Chat() botapi.Predicate {
	return func(c *botapi.Context) bool {
		ch, err := cctx.Chat(c)
		if err != nil {
			log.For(c.Bot.Logger()).Error(c, "no chat", log.Error(err))

			return false
		}

		return ch.ID != 0
	}
}

func Private() botapi.Predicate {
	return func(c *botapi.Context) bool {
		ch, err := cctx.Chat(c)
		if err != nil {
			log.For(c.Bot.Logger()).Error(c, "no chat", log.Error(err))

			return false
		}

		return ch.ID == 0
	}
}
