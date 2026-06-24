package stats

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

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
			ch.Lang, i18n.Cmd.AddNorm.ErrInvalidValue, i18n.CmdAddNormErrInvalidValueArgs(normValue, normMin, normMax)),
		)

		return err
	}

	name := GeneralNormName
	if len(args.Texts) > 0 && strings.TrimSpace(args.Texts[0]) != "" && strings.TrimSpace(args.Texts[0]) != h.DisplayNormName(ch.Lang, name) {
		name = args.Texts[0]
	}

	if !IsValidNormName(name) {
		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.AddNorm.ErrInvalidName, i18n.CmdAddNormErrInvalidNameArgs(name)))
		return err
	}

	if err := h.normRepository.Set(c.Context, ch.ID, name, normValue); err != nil {
		return fmt.Errorf("add norm: %w", err)
	}

	displayName := h.DisplayNormName(ch.Lang, name)
	_, err = c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.AddNorm.Added, i18n.CmdAddNormAddedArgs(displayName, normValue)))

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
					h.DisplayNormName(ch.Lang, n.Name),
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
		return fmt.Errorf("no args")
	}

	name := GeneralNormName

	if len(args.Texts) > 0 && strings.TrimSpace(args.Texts[0]) != "" {
		name = strings.TrimSpace(args.Texts[0])
	}
	norm, err := h.normRepository.Get(c.Context, ch.ID, name)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("show norm: %w", err)
		}

		if !IsValidNormName(name) {
			return nil
		}

		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.ShowNorm.NotFound,
			i18n.CmdShowNormNotFoundArgs(
				"<code>", "</code>",
				h.DisplayNormName(ch.Lang, name)),
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
				h.DisplayNormName(ch.Lang, norm.Name),
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
		return fmt.Errorf("no args")
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
			i18n.CmdDeleteNormErrNothingToDeleteArgs(h.DisplayNormName(ch.Lang, name))))
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
				h.DisplayNormName(ch.Lang, name),
			),
		),
	)

	return err
}
