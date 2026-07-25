package info

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/info/genshin"
	"activity-bot/internal/predicate"
	"bytes"
	"text/template"
)

type RenderData struct {
	Categories []RenderCategory
}

type RenderCategory struct {
	Emoji string
	Name  string
	Roles []RenderRole
}

type RenderRole struct {
	Name   string
	Status string
}

const rolesTemplate = `˚   𝘇 𐰁   𓆩 🗯 𓆪 ㅤ姿態哦   ૮ > . ა ✿ ꒱ . <tg-emoji emoji-id="5260536644913604662">👋</tg-emoji>

<blockquote><a href="http://t.me/HavenGateBot?start=true">бот для заявок</a></blockquote>
бронь – ✷
занятая роль – ♡゙
нуждаемся – !

{{range .Categories -}}
{{.Emoji}} {{.Name}}
<blockquote expandable>{{range .Roles}}{{.Name}} - {{.Status}}{{end}}
</blockquote>
{{end}}

менять роль можно только 2 или 3 раза для смены обратиться к владельцу или совладельцу
`

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

func BuildRoleStates(members []chatmember.ChatMember) map[string]RoleState {
	result := make(map[string]RoleState)

	roleIndex := make(map[string]struct{})

	for _, category := range genshin.Categories {
		for _, role := range category.Roles {
			roleIndex[predicate.NormalizeTag(role)] = struct{}{}
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

		status := "♡゙"

		result[tag] = RoleState{
			Role:   tag,
			Status: status,
		}
	}

	return result
}

func Render(states map[string]RoleState) (string, error) {
	data := RenderData{
		Categories: make([]RenderCategory, 0, len(genshin.Categories)),
	}

	for _, cat := range genshin.Categories {
		rc := RenderCategory{
			Emoji: cat.Emoji,
			Name:  cat.Name,
		}

		for _, role := range cat.Roles {
			status := ""

			if s, ok := states[predicate.NormalizeTag(role)]; ok {
				status = s.Status
			}

			rc.Roles = append(rc.Roles, RenderRole{
				Name:   role,
				Status: status,
			})
		}

		data.Categories = append(data.Categories, rc)
	}

	var buf bytes.Buffer

	if err := rolesTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
