package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
	"time"
)

func StatusTitle(status int16, count int) string {
	if count == 1 {
		switch status {
		case 1:
			return "Младший модератор"
		case 2:
			return "Старший модератор"
		case 3:
			return "Администратор"
		case 4:
			return "Совладелец"
		case 5:
			return "Владелец"
		default:
			return "Неизвестный"
		}
	}

	switch status {
	case 1:
		return "Младшие модераторы"
	case 2:
		return "Старшие модераторы"
	case 3:
		return "Администраторы"
	case 4:
		return "Совладельцы"
	case 5:
		return "Владельцы"
	default:
		return "Неизвестные"
	}
}
func FormatAdminsList(admins []model.ChatMember) string {
	var sb strings.Builder

	categories := map[int16][]model.ChatMember{}

	for _, a := range admins {
		if !a.LeftAt.IsZero() {
			continue
		}
		categories[a.Status] = append(categories[a.Status], a)
	}

	order := [5]int16{5, 4, 3, 2, 1}
	orderCustomEmojis := [5]int64{5258432055103991193, 5260268737738581704, 5260306546335690159, 5260274901016650206, 5260736442497247998}
	for i, status := range order {
		members := categories[status]
		if len(members) == 0 {
			continue
		}

		sb.WriteString("\n" + helpers.CustomEmoji(orderCustomEmojis[i], "⭐️") + " " + StatusTitle(status, len(members)) + ":\n")

		for i, m := range members {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, helpers.RoleEmojiLink(m)))
		}
	}

	return sb.String()
}

func FormatAdminAdded(user model.ChatMember, status int16) string {
	return fmt.Sprintf("Участнику %s %s ранг %d",
		helpers.RoleEmojiLink(user),
		helpers.Gendered(user.User.Gender, "назначен", "назначена"),
		status,
	)
}

func FormatAdminRemoved(user model.ChatMember) string {
	return fmt.Sprintf("%s %s из администраторов бота",
		helpers.RoleEmojiLink(user),
		helpers.Gendered(user.User.Gender, "удален", "удалена"),
	)
}

func FormatModerationAction(user model.ChatMember, action string, until *time.Time, reason string) string {
	var actionText string
	switch action {
	case "ban":
		actionText = helpers.Gendered(user.User.Gender, "забанен", "забанена")
	case "mute":
		actionText = helpers.Gendered(user.User.Gender, "замучен", "замучена")
	case "kick":
		actionText = helpers.Gendered(user.User.Gender, "был кикнут из чата", "была кикнута из чата")
	default:
		actionText = action
	}

	text := fmt.Sprintf("%s %s", helpers.RoleEmojiLink(user), actionText)

	if action != "kick" {
		if until != nil {
			text += fmt.Sprintf(" до %s", helpers.FormatToHumanDateTime(*until))
		} else {
			text += " навсегда"
		}
	}

	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	return text
}

func FormatWarnInfo(user model.ChatMember, count, maxWarns int, until *time.Time, reason string, banned bool) string {
	text := fmt.Sprintf("Участнику %s выдано предупреждение (%d/%d)", helpers.RoleEmojiLink(user), count, maxWarns)

	if until != nil {
		text += fmt.Sprintf(" до %s", helpers.FormatToHumanDateTime(*until))
	}

	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	if banned {
		text += fmt.Sprintf("\n\nУчастник %s за превышение лимита предупреждений.",
			helpers.Gendered(user.User.Gender, "забанен", "забанена"),
		)
	}

	return text
}

func FormatUnwarnInfo(user model.ChatMember, count, maxWarns int) string {
	return fmt.Sprintf("С участника %s снято предупреждение (%d/%d)", helpers.RoleEmojiLink(user), count, maxWarns)
}

func FormatWarnsCleared(user model.ChatMember) string {
	return fmt.Sprintf("Все предупреждения участника %s были аннулированы", helpers.RoleEmojiLink(user))
}

func FormatWarnlist(warns []model.Warn, maxWarns int) string {
	if len(warns) == 0 {
		return fmt.Sprintf("В этом чате нет активных предупреждений %s", helpers.SuccessEmoji())
	}

	var sb strings.Builder
	sb.WriteString("⚠️ <b>Список всех предупреждений в чате:</b>\n")

	userWarns := make(map[int64][]model.Warn)
	userOrder := make([]int64, 0)
	for _, w := range warns {
		if _, ok := userWarns[w.ChatMember.User.ID]; !ok {
			userOrder = append(userOrder, w.ChatMember.User.ID)
		}
		userWarns[w.ChatMember.User.ID] = append(userWarns[w.ChatMember.User.ID], w)
	}

	for _, userID := range userOrder {
		ws := userWarns[userID]
		sb.WriteString(fmt.Sprintf("\n👤 %s (активные: %d/%d):\n", helpers.RoleEmojiLink(ws[0].ChatMember), len(ws), maxWarns))
		for i, w := range ws {
			createdStr := helpers.FormatToHumanDateTime(w.CreatedAt)
			expireStr := ""
			if !w.ExpiresAt.IsZero() {
				expireStr = fmt.Sprintf(", истекает %s", helpers.FormatToHumanDateTime(w.ExpiresAt))
			}
			modName := helpers.RoleEmojiLink(w.Moderator)
			reasonStr := ""
			if w.Reason != "" {
				reasonStr = fmt.Sprintf(", причина: %s", w.Reason)
			}
			sb.WriteString(fmt.Sprintf("  %d. Выдан %s модератором %s%s%s\n",
				i+1, createdStr, modName, expireStr, reasonStr))
		}
	}

	return sb.String()
}

func FormatUnmuteInfo(user model.ChatMember) string {
	return fmt.Sprintf("Участник %s %s",
		helpers.RoleEmojiLink(user),
		helpers.Gendered(user.User.Gender, "размучен", "размучена"),
	)
}

func FormatDirectModerationAction(user model.ChatMember, chatTitle string, action string, until *time.Time, reason string) string {
	var actionText string
	switch action {
	case "ban":
		actionText = helpers.Gendered(user.User.Gender, "забанены", "забанены")
	case "kick":
		actionText = helpers.Gendered(user.User.Gender, "кикнуты", "кикнуты")
	default:
		actionText = action
	}

	text := fmt.Sprintf("Вы были %s в чате <b>%s</b>", actionText, chatTitle)

	if action == "ban" {
		if until != nil {
			text += fmt.Sprintf(" до %s", helpers.FormatToHumanDateTime(*until))
		} else {
			text += " навсегда"
		}
	}

	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	return text
}
