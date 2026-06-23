package stats

import (
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/telegram"
	"fmt"
	"log"
	"strconv"
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
	args := predicate.Args(c)
	if args == nil {
		return nil
	}

	text, ents := args.TextAndEntities()

	members := telegram.ParseMentionedMembers(c, h.chatMemberRepository, ch.ID, ents, text)

	for _, member := range members {
		fmt.Printf("%+v\n", member)
	}

	header := strings.TrimSpace(strings.Split(text, "\n")[0])

	fields := strings.Fields(header)
	if len(fields) == 0 {
		return fmt.Errorf("norm value is required")
	}

	normNumber, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return fmt.Errorf("parse norm number: %w", err)
	}

	var name string
	if len(fields) > 1 {
		name = strings.Join(fields[1:], " ")
	}
	name = strings.Join(strings.Fields(name), " ")

	if !IsValidNormName(name) {
		return nil
	}

	log.Printf("Norm: name=%s normNumber=%d", name, normNumber)
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
