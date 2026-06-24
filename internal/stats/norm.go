package stats

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gotd/botapi"
)

const GeneralNormName = "general"

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

	if len(args.Numbers) == 0 {
		return fmt.Errorf("add norm no number")
	}

	normValue := int32(args.Numbers[0])
	normMin := int32(1)
	normMax := int32(10000)
	if normValue < normMin || normValue > normMax {
		_, err := c.Reply(h.translator.TData(
			ch.Language, i18n.Cmd.AddNorm.ErrInvalidValue, i18n.CmdAddNormErrInvalidValueArgs(normMin, normMax)),
		)

		return err
	}

	name := GeneralNormName
	if len(args.Texts) > 0 && strings.TrimSpace(args.Texts[0]) != "" {
		name = args.Texts[0]
	}

	if len(args.Texts) > 0 && strings.TrimSpace(args.Texts[0]) != "" {
		name = args.Texts[0]
	}

	if !IsValidNormName(name) {
		return fmt.Errorf("invalid norm name")
	}

	if err := h.normRepository.Set(c.Context, ch.ID, name, normValue); err != nil {
		return fmt.Errorf("add norm: %w", err)
	}

	displayName := h.DisplayNormName(ch.Language, name)
	_, err = c.Reply(h.translator.TData(ch.Language, i18n.Cmd.AddNorm.Added, i18n.CmdAddNormAddedArgs(displayName, normValue)))

	return err
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

func (h *Handler) DisplayNormName(lang string, name string) string {
	switch name {
	case GeneralNormName:
		return h.translator.T(lang, i18n.Cmd.AddNorm.NormGeneral)
	default:
		return name
	}
}
