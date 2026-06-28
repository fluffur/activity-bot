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
	normValueStr := tghtml.Code(fmt.Sprintf("%d", normValue))

	if normValue < normMin || normValue > normMax {
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

	normID, err := h.repository.Set(c.Context, ch.ID, name, normValue)
	if err != nil {
		return fmt.Errorf("add norm: %w", err)
	}

	cms := args.Users
	userIDs := make([]int64, 0, len(cms))
	for _, cm := range cms {
		if cm.User.IsBot {
			continue
		}
		userIDs = append(userIDs, cm.User.ID)
	}

	if len(userIDs) > 0 {
		if err := h.repository.Assign(c.Context, normID, userIDs); err != nil {
			return fmt.Errorf("add norm assign: %w", err)
		}
	}

	displayName := LocalisedNormName(h.translator, ch.Lang, name)
	_, err = c.Reply(
		h.translator.TData(ch.Lang, i18n.Cmd.AddNorm.Added, i18n.CmdAddNormAddedArgs(tghtml.Bold(displayName), normValueStr)),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) ListNorms(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return err
	}

	norms, err := h.repository.List(c.Context, ch.ID)
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
	norm, err := h.repository.Get(c.Context, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("show norm: %w", err)
		}

		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.ShowNorm.NotFound,
			i18n.CmdShowNormNotFoundArgs(
				"<code>", "</code>",
				tghtml.Bold(LocalisedNormName(h.translator, ch.Lang, name))),
		),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	members, err := h.repository.GetNormMembers(c.Context, norm.ID)
	if err != nil {
		return fmt.Errorf("show norm get members: %w", err)
	}

	text := h.translator.TData(
		ch.Lang,
		i18n.Cmd.ShowNorm.Body,
		i18n.CmdShowNormBodyArgs(
			LocalisedNormName(h.translator, ch.Lang, norm.Name),
			norm.Value,
		),
	)

	if len(members) > 0 {
		text += "\n\n" + h.translator.T(ch.Lang, i18n.Cmd.ShowNorm.Members)

		for _, member := range members {
			text += "\n• " + tghtml.MemberLink(h.translator, ch, member)
		}
	} else {
		text += "\n\n" + h.translator.T(ch.Lang, i18n.Cmd.ShowNorm.AllMembers)
	}

	_, err = c.Reply(
		text,
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
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

	n, err := h.repository.Get(c.Context, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("delete norm: %w", err)
		}

		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.DeleteNorm.ErrNothingToDelete,
			i18n.CmdDeleteNormErrNothingToDeleteArgs(LocalisedNormName(h.translator, ch.Lang, name))))
		return err
	}

	if err := h.repository.Delete(c.Context, n.ID); err != nil {
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

func (h *Handler) AssignNorm(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("assign: %w", err)
	}

	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("assign: no args")
	}

	if len(args.Texts) == 0 {
		return fmt.Errorf("assign: no name")
	}

	name := args.Texts[0]

	n, err := h.repository.Get(c.Context, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("assign get norm: %w", err)
		}

		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.ShowNorm.NotFound,
			i18n.CmdShowNormNotFoundArgs(
				"<code>", "</code>",
				tghtml.Bold(LocalisedNormName(h.translator, ch.Lang, name))),
		), botapi.WithParseMode(botapi.ParseModeHTML))

		return err
	}

	cms := args.Users
	userIDs := make([]int64, 0, len(cms))
	for _, cm := range cms {
		if cm.User.IsBot && cm.IsLeft() {
			continue
		}
		userIDs = append(userIDs, cm.User.ID)
	}

	if len(userIDs) == 0 {
		_, err := c.Reply(h.translator.T(ch.Lang, i18n.Cmd.AssignNorm.NoUsers))
		return err
	}

	if err := h.repository.Assign(c.Context, n.ID, userIDs); err != nil {
		return fmt.Errorf("assign norm: %w", err)
	}

	_, err = c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.AssignNorm.Assigned,
		i18n.CmdAssignNormAssignedArgs(tghtml.Bold(LocalisedNormName(h.translator, ch.Lang, name)))),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) UnassignNorm(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("unassign: %w", err)
	}

	args, ok := predicate.GetParsedArgs(c)
	if !ok {
		return fmt.Errorf("unassign: no args")
	}

	if len(args.Texts) == 0 {
		return fmt.Errorf("unassign: no name")
	}

	name := args.Texts[0]

	n, err := h.repository.Get(c.Context, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unassign get norm: %w", err)
		}

		_, err := c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.ShowNorm.NotFound,
			i18n.CmdShowNormNotFoundArgs(
				"<code>", "</code>",
				tghtml.Bold(LocalisedNormName(h.translator, ch.Lang, name))),
		), botapi.WithParseMode(botapi.ParseModeHTML))

		return err
	}

	cms := args.Users
	userIDs := make([]int64, 0, len(cms))
	for _, cm := range cms {
		if cm.User.IsBot {
			continue
		}
		userIDs = append(userIDs, cm.User.ID)
	}

	if len(userIDs) == 0 {
		_, err := c.Reply(h.translator.T(ch.Lang, i18n.Cmd.UnassignNorm.NoUsers))
		return err
	}

	if err := h.repository.Unassign(c.Context, n.ID, userIDs); err != nil {
		return fmt.Errorf("unassign norm: %w", err)
	}

	_, err = c.Reply(h.translator.TData(ch.Lang, i18n.Cmd.UnassignNorm.Unassigned,
		i18n.CmdUnassignNormUnassignedArgs(tghtml.Bold(LocalisedNormName(h.translator, ch.Lang, name)))),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err

}
