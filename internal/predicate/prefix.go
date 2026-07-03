package predicate

import (
	"activity-bot/internal/cctx"
	"strings"

	"github.com/gotd/botapi"
)

func SensitiveCommand() botapi.Predicate {
	return func(c *botapi.Context) bool {
		msg := c.Message()
		if msg == nil {
			return false
		}
		p := cctx.MustCommandPrefix(c)

		return p != "" || strings.HasPrefix(msg.Text, "+")
	}
}
