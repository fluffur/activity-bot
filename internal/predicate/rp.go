package predicate

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/rp"

	"github.com/gotd/botapi"
)

func RPCommand(repo rp.Repository, prefixes []string) botapi.Predicate {
	return func(c *botapi.Context) bool {
		parsed, ok := ParseMessage(c, prefixes)
		if !ok {
			return false
		}

		definition, triggerLen, err := repo.Match(
			c,
			cctx.MustChat(c).ID,
			parsed.TrimmedText,
		)
		if err != nil {
			return false
		}

		args := BuildArgsMessage(
			c.Message(),
			parsed.Text,
			parsed.Entities,
			parsed.Prefix,
			parsed.LeadingSpacesRunes,
			triggerLen,
		)

		c.Context = cctx.WithCommandPrefix(c.Context, parsed.Prefix)
		c.Context = cctx.WithArgsMessage(c.Context, args)
		c.Context = cctx.WithRPCommand(c.Context, definition)

		return true
	}
}
