package rolepost

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/predicate"
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

	b.WriteString(`˚   𝘇 𐰁   𓆩 🗯 𓆪 ㅤ姿態哦   ૮ > . ა ✿ ꒱ .

<blockquote><a href="http://t.me/HavenGateBot?start=true">бот для заявок</a></blockquote>
бронь – ✷
занятая роль – ♡゙
нуждаемся – !
`)

	for _, category := range Categories {

		b.WriteString("\n\n")

		b.WriteString(category.Emoji)
		b.WriteString(" ")
		b.WriteString(category.Name)
		b.WriteString("\n")

		for _, role := range category.Roles {

			b.WriteString(role)
			b.WriteString(" - ")

			if state, ok := roles[role]; ok {
				b.WriteString(state.Status)
			}

			b.WriteString("\n")
		}

		//b.WriteString("</blockquote>")
	}

	b.WriteString(`

менять роль можно только 2 или 3 раза для смены обратиться к владельцу или совладельцу
`)

	return b.String()
}
