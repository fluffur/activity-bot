package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
	"time"
)

func FormatAdminsList(admins []model.ChatMember) string {
	var sb strings.Builder
	sb.WriteString("👮 Администраторы бота:\n")
	for i, a := range admins {
		sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, helpers.RoleLink(a)))
	}
	return sb.String()
}

func FormatAdminAdded(user model.ChatMember) string {
	return fmt.Sprintf("Участник %s %s администратором бота",
		helpers.RoleLink(user),
		helpers.Gendered(user.User.Gender, "назначен", "назначена"),
	)
}

func FormatAdminRemoved(user model.ChatMember) string {
	return fmt.Sprintf("%s %s из администраторов бота",
		helpers.RoleLink(user),
		helpers.Gendered(user.User.Gender, "удален", "удалена"),
	)
}

func FormatDevelopersList(users []model.User, roles []string) string {
	var sb strings.Builder
	sb.WriteString("🛠 Разработчики бота:\n")
	for i, u := range users {
		sb.WriteString(fmt.Sprintf("\n%d. %s (%s)", i+1, helpers.UserLink(u), roles[i]))
	}
	return sb.String()
}

func FormatDeveloperAdded(user model.User, role string) string {
	return fmt.Sprintf("Участник %s %s разработчиком бота с ролью %s",
		helpers.UserLink(user),
		helpers.Gendered(user.Gender, "назначен", "назначена"),
		role,
	)
}

func FormatDeveloperRemoved(user model.User) string {
	return fmt.Sprintf("Участник %s %s из списка разработчиков",
		helpers.UserLink(user),
		helpers.Gendered(user.Gender, "удален", "удалена"),
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

	text := fmt.Sprintf("%s %s", helpers.RoleLink(user), actionText)

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
	text := fmt.Sprintf("Участнику %s выдано предупреждение (%d/%d)", helpers.RoleLink(user), count, maxWarns)

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
	return fmt.Sprintf("С участника %s снято предупреждение (%d/%d)", helpers.RoleLink(user), count, maxWarns)
}

func FormatWarnsCleared(user model.ChatMember) string {
	return fmt.Sprintf("Все предупреждения участника %s были аннулированы", helpers.RoleLink(user))
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
		sb.WriteString(fmt.Sprintf("\n👤 %s (активные: %d/%d):\n", helpers.RoleLink(ws[0].ChatMember), len(ws), maxWarns))
		for i, w := range ws {
			createdStr := helpers.FormatToHumanDateTime(w.CreatedAt)
			expireStr := ""
			if !w.ExpiresAt.IsZero() {
				expireStr = fmt.Sprintf(", истекает %s", helpers.FormatToHumanDateTime(w.ExpiresAt))
			}
			modName := helpers.RoleLink(w.Moderator)
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
		helpers.RoleLink(user),
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
