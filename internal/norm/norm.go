package norm

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

const GeneralNormName = "general"

func (h *Handler) AddNorm(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	args := cctx.MustArgs(c)

	if len(args.Numbers) == 0 {
		return fmt.Errorf("add norm: no number")
	}

	value := int32(args.Numbers[0])

	const (
		min = int32(1)
		max = int32(10000)
	)

	if value < min || value > max {
		_, err := c.Reply(
			loc.T(i18n.Cmd.AddNorm.ErrInvalidValue, i18n.CmdAddNormErrInvalidValueData{
				Value: tghtml.Code(fmt.Sprintf("%d", value)),
				Min:   min,
				Max:   max,
			}),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)
		return err
	}

	name := GeneralNormName
	if len(args.Texts) > 0 {
		if text := strings.TrimSpace(args.Texts[0]); text != "" && text != LocalisedNormName(loc, name) {
			name = text
		}
	}

	normID, err := h.repository.Set(c, ch.ID, name, value)
	if err != nil {
		return fmt.Errorf("add norm: %w", err)
	}

	var ids []int64
	for _, cm := range args.Users {
		if !cm.User.IsBot {
			ids = append(ids, cm.User.ID)
		}
	}

	if len(ids) > 0 {
		if err := h.repository.Assign(c, normID, ids); err != nil {
			return fmt.Errorf("assign norm: %w", err)
		}
	}

	_, err = c.Reply(
		loc.T(i18n.Cmd.AddNorm.Added, i18n.CmdAddNormAddedData{
			Name:  tghtml.Bold(LocalisedNormName(loc, name)),
			Value: tghtml.Code(fmt.Sprintf("%d", value)),
		}),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) ListNorms(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)

	norms, err := h.repository.List(c, ch.ID)
	if err != nil {
		return err
	}

	if len(norms) == 0 {
		_, err := c.Reply(loc.T(i18n.Cmd.ListNorms.Empty, nil))
		return err
	}

	var b strings.Builder

	b.WriteString(loc.T(i18n.Cmd.ListNorms.Title, nil))
	b.WriteString("\n\n")

	for _, n := range norms {
		b.WriteString(loc.T(i18n.Cmd.ListNorms.Item, i18n.CmdListNormsItemData{
			Name:  LocalisedNormName(loc, n.Name),
			Value: n.Value,
		}))
		b.WriteByte('\n')
	}

	_, err = c.Reply(b.String())
	return err
}

func (h *Handler) ShowNorm(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	args := cctx.MustArgs(c)

	name := GeneralNormName
	if len(args.Texts) > 0 {
		if text := strings.TrimSpace(args.Texts[0]); text != "" {
			name = text
		}
	}

	n, err := h.repository.Get(c, ch.ID, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	members, err := h.repository.GetNormMembers(c, n.ID)
	if err != nil {
		return err
	}

	var b strings.Builder

	b.WriteString(loc.T(i18n.Cmd.ShowNorm.Body, i18n.CmdShowNormBodyData{
		Name:  LocalisedNormName(loc, n.Name),
		Value: n.Value,
	}))

	b.WriteString("\n\n")

	if len(members) == 0 {
		b.WriteString(loc.T(i18n.Cmd.ShowNorm.AllMembers, nil))
	} else {
		b.WriteString(loc.T(i18n.Cmd.ShowNorm.Members, nil))
		for _, m := range members {
			b.WriteString("\n• ")
			b.WriteString(tghtml.MemberLink(loc, ch, m))
		}
	}

	_, err = c.Reply(
		b.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func (h *Handler) DeleteNorm(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	args := cctx.MustArgs(c)

	if len(args.Texts) == 0 {
		return fmt.Errorf("delete norm: no name")
	}

	name := strings.TrimSpace(args.Texts[0])

	if name == loc.T(i18n.Cmd.AddNorm.NormGeneral, nil) {
		name = GeneralNormName
	}

	n, err := h.repository.Get(c, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		_, err := c.Reply(loc.T(i18n.Cmd.DeleteNorm.ErrNothingToDelete, i18n.CmdDeleteNormErrNothingToDeleteData{
			Name: LocalisedNormName(loc, name),
		}))
		return err
	}

	if err := h.repository.Delete(c, n.ID); err != nil {
		return err
	}

	_, err = c.Reply(loc.T(i18n.Cmd.DeleteNorm.Deleted, i18n.CmdDeleteNormDeletedData{
		Name: LocalisedNormName(loc, name),
	}))

	return err
}
func (h *Handler) AssignNorm(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	args := cctx.MustArgs(c)

	if len(args.Texts) == 0 {
		return fmt.Errorf("assign: no name")
	}

	name := strings.TrimSpace(args.Texts[0])

	n, err := h.repository.Get(c, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("assign get norm: %w", err)
		}

		_, err = c.Reply(
			loc.T(i18n.Cmd.ShowNorm.NotFound, i18n.CmdShowNormNotFoundData{
				Name:    tghtml.Bold(LocalisedNormName(loc, name)),
				Code:    "<code>",
				CodeEnd: "</code>",
			}),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	var userIDs []int64
	for _, cm := range args.Users {
		if cm.User.IsBot && cm.IsLeft() {
			continue
		}
		userIDs = append(userIDs, cm.User.ID)
	}

	if len(userIDs) == 0 {
		_, err = c.Reply(loc.T(i18n.Cmd.AssignNorm.NoUsers, nil))
		return err
	}

	if err := h.repository.Assign(c, n.ID, userIDs); err != nil {
		return fmt.Errorf("assign norm: %w", err)
	}

	_, err = c.Reply(
		loc.T(i18n.Cmd.AssignNorm.Assigned, i18n.CmdAssignNormAssignedData{
			Name: tghtml.Bold(LocalisedNormName(loc, name)),
		}),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}

func (h *Handler) UnassignNorm(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	args := cctx.MustArgs(c)

	if len(args.Texts) == 0 {
		return fmt.Errorf("unassign: no name")
	}

	name := strings.TrimSpace(args.Texts[0])

	n, err := h.repository.Get(c, ch.ID, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unassign get norm: %w", err)
		}

		_, err = c.Reply(
			loc.T(i18n.Cmd.ShowNorm.NotFound, i18n.CmdShowNormNotFoundData{
				Name:    tghtml.Bold(LocalisedNormName(loc, name)),
				Code:    "<code>",
				CodeEnd: "</code>",
			}),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	}

	var userIDs []int64
	for _, cm := range args.Users {
		if cm.User.IsBot {
			continue
		}
		userIDs = append(userIDs, cm.User.ID)
	}

	if len(userIDs) == 0 {
		_, err = c.Reply(loc.T(i18n.Cmd.UnassignNorm.NoUsers, nil))
		return err
	}

	if err := h.repository.Unassign(c, n.ID, userIDs); err != nil {
		return fmt.Errorf("unassign norm: %w", err)
	}

	_, err = c.Reply(
		loc.T(i18n.Cmd.UnassignNorm.Unassigned, i18n.CmdUnassignNormUnassignedData{
			Name: tghtml.Bold(LocalisedNormName(loc, name)),
		}),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}
