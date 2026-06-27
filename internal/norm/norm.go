package norm

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/tghtml"
	"database/sql"
	"errors"
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

	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("add norm: no args")
	}

	if len(args.Numbers) == 0 {
		return fmt.Errorf("add norm no number")
	}

	normValue := int32(args.Numbers[0])
	normMin := int32(1)
	normMax := int32(10000)
	if normValue < normMin || normValue > normMax {
		normValueStr := tghtml.Bold(fmt.Sprintf("%d", normValue))
		_, err := c.Reply(h.translator.TData(
			ch.Lang, i18n.Cmd.AddNorm.ErrInvalidValue, i18n.CmdAddNormErrInvalidValueArgs(normMax, normMin, normValueStr)),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	name := GeneralNormName
	if len(args.Texts) > 0 {
		text := strings.TrimSpace(args.Texts[0])

		if text != "" && text != LocalisedNormName(h.translator, ch.Lang, name) {
			name = text
		}
	}

	if err := h.normRepository.Set(c.Context, ch.ID, name, normValue); err != nil {
		return fmt.Errorf("add norm: %w", err)
	}

	displayName := LocalisedNormName(h.translator, ch.Lang, name)
	_, err = c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.AddNorm.Added, i18n.CmdAddNormAddedArgs(displayName, normValue)))

	return err
}

func (h *Handler) ListNorms(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return err
	}

	norms, err := h.normRepository.List(c.Context, ch.ID)
	if err != nil {
		return err
	}

	if len(norms) == 0 {
		_, err = c.Reply(
			h.translator.T(ch.Lang, i18n.Cmd.ListNorms.Empty),
		)
		return err
	}

	var b strings.Builder

	b.WriteString(h.translator.T(ch.Lang, i18n.Cmd.ListNorms.Title))
	b.WriteString("\n\n")

	for _, n := range norms {
		b.WriteString(
			h.translator.TData(
				ch.Lang,
				i18n.Cmd.ListNorms.Item,
				i18n.CmdListNormsItemArgs(
					LocalisedNormName(h.translator, ch.Lang, n.Name),
					n.Value,
				),
			),
		)

		b.WriteString("\n")
	}
	_, err = c.Reply(b.String())

	return err
}

func (h *Handler) ShowNorm(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return err
	}

	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("show norm: no args")
	}

	name := GeneralNormName

	if len(args.Texts) > 0 && strings.TrimSpace(args.Texts[0]) != "" {
		name = strings.TrimSpace(args.Texts[0])
	}
	norm, err := h.normRepository.Get(c.Context, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("show norm: %w", err)
		}

		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.ShowNorm.NotFound,
			i18n.CmdShowNormNotFoundArgs(
				"<code>", "</code>",
				LocalisedNormName(h.translator, ch.Lang, name)),
		),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	_, err = c.Reply(
		h.translator.TData(
			ch.Lang,
			i18n.Cmd.ShowNorm.Body,
			i18n.CmdShowNormBodyArgs(
				LocalisedNormName(h.translator, ch.Lang, norm.Name),
				norm.Value,
			),
		),
	)

	return err
}

func (h *Handler) DeleteNorm(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return err
	}

	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("delete norm: no args")
	}
	if len(args.Texts) == 0 {
		return fmt.Errorf("delete norm: no name")
	}

	name := strings.TrimSpace(args.Texts[0])

	if name == h.translator.T(ch.Lang, i18n.Cmd.AddNorm.NormGeneral) {
		name = GeneralNormName
	}

	if _, err := h.normRepository.Get(c.Context, ch.ID, name); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("delete norm: %w", err)
		}

		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.DeleteNorm.ErrNothingToDelete,
			i18n.CmdDeleteNormErrNothingToDeleteArgs(LocalisedNormName(h.translator, ch.Lang, name))))
		return err
	}

	if err := h.normRepository.Delete(c.Context, ch.ID, name); err != nil {
		return fmt.Errorf("delete norm: %w", err)
	}

	_, err = c.Reply(
		h.translator.TData(
			ch.Lang,
			i18n.Cmd.DeleteNorm.Deleted,
			i18n.CmdDeleteNormDeletedArgs(
				LocalisedNormName(h.translator, ch.Lang, name),
			),
		),
	)

	return err
}

func isValidNormName(name string) bool {
	name = strings.TrimSpace(name)

	if name == "" {
		return false
	}

	length := utf8.RuneCountInString(name)

	if length > 20 || length < 2 {
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

func LocalisedNormName(t *i18n.Translator, lang string, name string) string {
	switch name {
	case GeneralNormName:
		return t.T(lang, i18n.Cmd.AddNorm.NormGeneral)
	default:
		return name
	}
}
