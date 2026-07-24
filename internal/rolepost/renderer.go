package rolepost

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/predicate"
	"bytes"
	"html/template"
	"strings"
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

	for _, category := range Categories {
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

func Render(roles map[string]RoleState) string {

	var b strings.Builder

	b.WriteString(`˚   𝘇 𐰁   𓆩 🗯 𓆪 ㅤ姿態哦   ૮ > . ა ✿ ꒱ . <tg-emoji emoji-id="5260536644913604662">👋</tg-emoji>

<blockquote><a href="http://t.me/HavenGateBot?start=true">бот для заявок</a></blockquote>
бронь – ✷
занятая роль – ♡゙
нуждаемся – !
`)

	for _, category := range Categories {

		b.WriteString("\n\n<blockquote expandable>")

		b.WriteString(category.Emoji)
		b.WriteString(" ")
		b.WriteString(category.Name)
		b.WriteString("\n")

		for _, role := range category.Roles {

			b.WriteString(role)
			b.WriteString(" - ")

			if state, ok := roles[predicate.NormalizeTag(role)]; ok {
				b.WriteString(state.Status)
			}

			b.WriteString("\n")
		}

		b.WriteString("</blockquote>")
	}

	b.WriteString(`

менять роль можно только 2 или 3 раза для смены обратиться к владельцу или совладельцу
`)

	return b.String()
}

const applicationTemplate = `
──────────────────────────────────────
               <tg-emoji emoji-id="5404728034299239682">🔠</tg-emoji><tg-emoji emoji-id="5402092367488508573">🔠</tg-emoji><tg-emoji emoji-id="5402092367488508573">🔠</tg-emoji><tg-emoji emoji-id="5402336351695695846">🔠</tg-emoji><tg-emoji emoji-id="5402424050632909510">🔠</tg-emoji>
──────────────────────────────────────

     статус – <u>{{if .Open}}открыт{{else}}закрыт{{end}}</u>

                {{.Current}}/{{.Max}}

     для вступления

     <a href="http://t.me/HavenGateBot?start=true">напишите свою роль боту</a>

      ˚₊·— ⁠⁠⁠♡ <a href="https://t.me/H4venflood/15">список ролей</a>

˚₊‧✩ ˚₊‧꒰ა ʚིᵋº‌‌‌‌‌‌ᵌɞྀ ໒꒱ ‧₊˚ ✩‧₊˚•*¨*•.¸¸☆*
`

var applicationTmpl = template.Must(
	template.New("application").Parse(applicationTemplate),
)

type ApplicationPostData struct {
	Current int
	Max     int
	Open    bool
}

func RenderApplication(current, max int) (string, error) {
	var b bytes.Buffer

	err := applicationTmpl.Execute(&b, ApplicationPostData{
		Current: current,
		Max:     max,
		Open:    current < max,
	})

	if err != nil {
		return "", err
	}

	return b.String(), nil
}
