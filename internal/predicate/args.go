package predicate

import (
	"strings"

	"github.com/gotd/botapi"
)

func NoArgs() botapi.Predicate {
	return func(c *botapi.Context) bool {
		args := Args(c)

		return strings.TrimSpace(args.Text) == ""
	}
}
