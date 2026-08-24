package roles

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/info/genshin"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/predicate"
	"activity-bot/internal/rule"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/gotd/botapi"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

const RolesCategory command.Category = "roles"

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"rolesgenshin",
			h.ImportGenshin,
			i18n.Cmd.Roles.Genshin.Desc,
			RolesCategory,
			option.WithAliases("роли геншин"),
			option.WithPermission(permission.StatusCoOwner),
		),
		action.NewCommand(
			"reserve",
			h.ReserveRole,
			i18n.Cmd.Roles.Reserve.Desc,
			RolesCategory,
			option.WithAliases("бронь"),
			option.WithRules(rule.User().Optional(), rule.Text()),
			option.WithPermission(permission.StatusAdmin),
		),
	}
}

func (h *Handler) ReserveRole(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	text, ok := cctx.MustArgs(c).Text()
	if !ok {
		return nil
	}
	spew.Dump(cctx.MustArgs(c))
	role, err := h.repo.GetRoleByNameOrAlias(c, ch.ID, "Genshin Impact", predicate.NormalizeTag(text))

	if err != nil {
		return fmt.Errorf("reserve role: %w", err)
	}

	if err := h.repo.CreateRoleReservation(c, ch.ID, 0, role.ID); err != nil {
		return fmt.Errorf("create role reservation: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	_, err = c.Reply(loc.T(i18n.Cmd.Roles.Reserve.Success, nil))

	return err
}

func (h *Handler) ImportGenshin(c *botapi.Context) error {
	ch := cctx.MustChat(c)

	if err := h.repo.CreateRoleTemplate(c, ch.ID, "Genshin Impact", GenshinRoles()); err != nil {
		return fmt.Errorf("create role template: %w", err)
	}

	loc := cctx.MustLocalizer(c)

	_, err := c.Reply(loc.T(i18n.Cmd.Roles.Genshin.Success, nil))
	return err
}

func GenshinRoles() []Category {
	result := make([]Category, 0, len(genshin.Categories))

	for _, category := range genshin.Categories {
		result = append(result, Category{
			Name:  category.Name,
			Roles: mapRoles(category.Roles),
		})
	}

	return result
}

func mapRoles(source []genshin.Role) []Role {
	result := make([]Role, 0, len(source))

	for _, role := range source {
		result = append(result, Role{
			Name:    role.Name,
			Emoji:   role.Emoji,
			Aliases: role.Aliases,
		})
	}

	return result
}
