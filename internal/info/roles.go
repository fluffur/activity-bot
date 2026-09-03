package info

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/predicate"
	"activity-bot/internal/roles"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

type RenderData struct {
	Categories []RenderCategory
	Header     string
	Footer     string
}

type RenderCategory struct {
	Emoji string
	Name  string
	Roles []RenderRole
	Text  string
}

type RenderRole struct {
	Name   string
	Status string
}

const rolesTemplate = `{{.Header}}
{{range .Categories -}}
{{.Name}}
<blockquote expandable>{{.Text}}</blockquote>

{{end}}`

var rolesTmpl = template.Must(
	template.New("roles").Parse(rolesTemplate),
)

type RoleState struct {
	Role string

	// пусто
	// "♡゙"
	// "✷"
	// "!"
	Status string
}

func BuildRoleStates(
	fandoms []roles.Fandom,
	members []chatmember.ChatMember,
	reservations []roles.RoleReservation,
) map[string]RoleState {
	result := make(map[string]RoleState)

	roleIndex := make(map[string]struct{})

	for _, f := range fandoms {
		for _, category := range f.Categories {
			for _, role := range category.Roles {
				roleIndex[predicate.NormalizeTag(role.Name)] = struct{}{}
			}
		}
	}

	for _, reservation := range reservations {
		tag := predicate.NormalizeTag(reservation.Role.Name)

		if tag == "" {
			continue
		}

		if _, ok := roleIndex[tag]; !ok {
			continue
		}

		result[tag] = RoleState{
			Role:   tag,
			Status: "✷",
		}
	}

	for _, member := range members {
		tag := predicate.NormalizeTag(member.Tag)

		if tag == "" {
			continue
		}

		if _, ok := roleIndex[tag]; !ok {
			continue
		}

		result[tag] = RoleState{
			Role:   tag,
			Status: "♡゙",
		}
	}

	return result
}

func formatRole(role RenderRole) string {
	return fmt.Sprintf("%s  — %s", role.Name, role.Status)
}

func formatRoles(roles []RenderRole) string {
	var buf bytes.Buffer

	for i := 0; i < len(roles); i += 2 {
		left := formatRole(roles[i])

		buf.WriteString(left)

		if i+1 < len(roles) {
			right := formatRole(roles[i+1])

			padding := 20 - len([]rune(left))

			if padding < 4 {
				padding = 4
			}

			buf.WriteString(strings.Repeat(" ", padding))
			buf.WriteString(right)
		}

		if i+2 < len(roles) {
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

func Render(
	fandoms []roles.Fandom,
	states map[string]RoleState,
) ([]string, error) {
	if len(fandoms) != 1 {
		return nil, errors.New("must provide exactly one random role")
	}

	categories := make([]RenderCategory, 0, len(fandoms[0].Categories))

	for _, cat := range fandoms[0].Categories {
		rc := RenderCategory{
			Name: formatCategoryName(cat.Name),
		}

		for _, role := range cat.Roles {
			status := ""

			if s, ok := states[predicate.NormalizeTag(role.Name)]; ok {
				status = s.Status
			}

			rc.Roles = append(rc.Roles, RenderRole{
				Name:   role.Name,
				Status: status,
			})
		}

		rc.Text = formatRoles(rc.Roles)

		categories = append(categories, rc)
	}

	header := `˚   𝘇 𐰁   𓆩 🗯 𓆪 ㅤ姿態哦   ૮ > . ა ✿ ꒱ . 

<blockquote><a href="https://t.me/HavenGateBot?start=true">бот для заявок</a></blockquote>
бронь – ✷
занятая роль – ♡゙
нуждаемся – !
`

	firstCount := 3

	if firstCount > len(categories) {
		firstCount = len(categories)
	}

	post1, err := renderCategories(categories[:firstCount], header)
	if err != nil {
		return nil, err
	}

	post2, err := renderCategories(categories[firstCount:], "")
	if err != nil {
		return nil, err
	}

	return []string{post1, post2}, nil
}

func renderCategories(
	categories []RenderCategory,
	header string,
) (string, error) {
	data := RenderData{
		Categories: categories,
		Header:     header,
	}

	var buf bytes.Buffer

	if err := rolesTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func formatCategoryName(name string) string {
	return fmt.Sprintf(
		"    ˖ ݁𖥔.  %s  .𖥔 ݁ ˖",
		name,
	)
}
