package stats

import (
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"fmt"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gotd/botapi"
)

func (h *Handler) AddNorm(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("add norm: %w", err)
	}
	_ = ch

	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("no args")
	}

	log.Printf("adding norm args: %+v", args)

	return nil
}

func IsValidNormName(name string) bool {
	name = strings.TrimSpace(name)

	if name == "" {
		return false
	}

	if utf8.RuneCountInString(name) > 64 {
		return false
	}

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			continue
		}

		return false
	}

	return true
}
