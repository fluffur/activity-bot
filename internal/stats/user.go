package stats

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) Profile(c *botapi.Context) error {
	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("profile invalid argument")
	}

	var u chatmember.ChatMember
	if len(args.Users) != 0 {
		u = args.Users[0]
	} else {
		us, err := cctx.ChatMember(c.Context)
		if err != nil {
			return fmt.Errorf("profile: %w", err)
		}
		u = us
	}
	_ = u

	return nil
}
