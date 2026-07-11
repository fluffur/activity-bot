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
		if chatID, ok := c.Chat(); ok {
			if id, ok := ChatIDInt64(chatID); ok {
				return ch.ID == 0 || id != ch.ID
			}
		}

		return ch.ID == 0
	}
}

func ChatIDInt64(id botapi.ChatID) (int64, bool) {
	v, ok := id.(botapi.ChatIDInt)
	if !ok {
		return 0, false
	}

	return int64(v), true
}
