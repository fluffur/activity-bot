package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
	"time"
)

func FormatAdminsList(admins []model.User) string {
	var sb strings.Builder
	sb.WriteString("👮 Администраторы бота:\n")
	for i, a := range admins {
		sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, helpers.Link(a)))
	}
	return sb.String()
}

func FormatAdminAdded(user model.User) string {
	return fmt.Sprintf("Участник %s %s администратором бота",
		helpers.Link(user),
		helpers.Gendered(user.Gender, "назначен", "назначена", "назначен(а)"),
	)
}

func FormatAdminRemoved(user model.User) string {
	return fmt.Sprintf("Участник %s %s из администраторов бота",
		helpers.Link(user),
		helpers.Gendered(user.Gender, "удалён", "удалена", "удалён(а)"),
	)
}

func FormatDevelopersList(users []model.User, levels []int16) string {
	var sb strings.Builder
	sb.WriteString("🛠 Разработчики бота:\n")
	mapping := map[int16]string{
		0: "участник",
		3: "администратор",
		5: "создатель",
	}
	for i, u := range users {
		roleName := mapping[levels[i]]
		if roleName == "" {
			roleName = fmt.Sprintf("уровень %d", levels[i])
		}
		sb.WriteString(fmt.Sprintf("\n%d. %s (%s)", i+1, helpers.Link(u), roleName))
	}
	return sb.String()
}

func FormatDeveloperAdded(user model.User, level int16) string {
	mapping := map[int16]string{
		0: "участник",
		3: "администратор",
		5: "создатель",
	}
	roleName := mapping[level]
	if roleName == "" {
		roleName = fmt.Sprintf("уровень %d", level)
	}
	return fmt.Sprintf("Участник %s %s разработчиком бота с ролью %s",
		helpers.Link(user),
		helpers.Gendered(user.Gender, "назначен", "назначена", "назначен(а)"),
		roleName,
	)
}

func FormatDeveloperRemoved(user model.User) string {
	return fmt.Sprintf("Участник %s %s из списка разработчиков",
		helpers.Link(user),
		helpers.Gendered(user.Gender, "удален", "удалена", "удален(а)"),
	)
}

func FormatModerationAction(user model.User, action string, until *time.Time, reason string) string {
	var actionText string
	switch action {
	case "ban":
		actionText = helpers.Gendered(user.Gender, "забанен", "забанена", "забанен(а)")
	case "mute":
		actionText = helpers.Gendered(user.Gender, "замучен", "замучена", "замучен(а)")
	case "kick":
		actionText = helpers.Gendered(user.Gender,
			"был кикнут из чата",
			"была кикнута из чата",
			"кикнут(а) из чата",
		)
	default:
		actionText = action
	}

	text := fmt.Sprintf("Пользователь %s %s", helpers.Link(user), actionText)

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

func FormatWarnInfo(user model.User, count, maxWarns int, until *time.Time, reason string, banned bool) string {
	text := fmt.Sprintf("Пользователю %s выдано предупреждение (%d/%d)", helpers.Link(user), count, maxWarns)

	if until != nil {
		text += fmt.Sprintf(" до %s", helpers.FormatToHumanDateTime(*until))
	}

	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	if banned {
		text += fmt.Sprintf("\n\nПользователь %s за превышение лимита предупреждений.",
			helpers.Gendered(user.Gender, "забанен", "забанена", "забанен(а)"),
		)
	}

	return text
}

func FormatUnwarnInfo(user model.User, count, maxWarns int) string {
	return fmt.Sprintf("С пользователя %s снято предупреждение (%d/%d)", helpers.Link(user), count, maxWarns)
}

func FormatWarnsCleared(user model.User) string {
	return fmt.Sprintf("Все предупреждения пользователя %s были аннулированы", helpers.Link(user))
}

func FormatWarnlist(warns []model.Warn, maxWarns int) string {
	if len(warns) == 0 {
		return "В этом чате нет активных предупреждений ✅"
	}

	var sb strings.Builder
	sb.WriteString("⚠️ <b>Список всех предупреждений в чате:</b>\n")

	userWarns := make(map[int64][]model.Warn)
	userOrder := make([]int64, 0)
	for _, w := range warns {
		if _, ok := userWarns[w.User.ID]; !ok {
			userOrder = append(userOrder, w.User.ID)
		}
		userWarns[w.User.ID] = append(userWarns[w.User.ID], w)
	}

	for _, userID := range userOrder {
		ws := userWarns[userID]
		u := ws[0].User
		sb.WriteString(fmt.Sprintf("\n👤 %s (активные: %d/%d):\n", helpers.Link(u), len(ws), maxWarns))
		for i, w := range ws {
			createdStr := helpers.FormatToHumanDateTime(w.CreatedAt)
			expireStr := ""
			if !w.ExpiresAt.IsZero() {
				expireStr = fmt.Sprintf(", истекает %s", helpers.FormatToHumanDateTime(w.ExpiresAt))
			}
			modName := helpers.Link(w.Moderator)
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

func FormatUnmuteInfo(user model.User) string {
	return fmt.Sprintf("Пользователь %s %s",
		helpers.Link(user),
		helpers.Gendered(user.Gender, "размучен", "размучена", "размучен(а)"),
	)
}

func FormatDirectModerationAction(user model.User, chatTitle string, action string, until *time.Time, reason string) string {
	var actionText string
	switch action {
	case "ban":
		actionText = helpers.Gendered(user.Gender, "забанены", "забанены", "забанены")
	case "kick":
		actionText = helpers.Gendered(user.Gender, "кикнуты", "кикнуты", "кикнуты")
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
