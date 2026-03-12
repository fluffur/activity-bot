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
		title := a.CustomTitle
		if a.CustomTitle == "" {
			title = a.User.FirstName
		}
		sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, helpers.LinkWithContent(a.User, title)))
	}
	return sb.String()
}

func FormatAdminAdded(user model.User) string {
	return fmt.Sprintf("Участник %s %s администратором бота",
		helpers.UserLink(user),
		helpers.Gendered(user.Gender, "назначен", "назначена"),
	)
}

func FormatAdminRemoved(user model.User) string {
	return fmt.Sprintf("Участник %s %s из администраторов бота",
		helpers.UserLink(user),
		helpers.Gendered(user.Gender, "удалён", "удалена"),
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

func FormatModerationAction(user model.User, action string, until *time.Time, reason string) string {
	var actionText string
	switch action {
	case "ban":
		actionText = helpers.Gendered(user.Gender, "забанен", "забанена")
	case "mute":
		actionText = helpers.Gendered(user.Gender, "замучен", "замучена")
	case "kick":
		actionText = helpers.Gendered(user.Gender, "был кикнут из чата", "была кикнута из чата")
	default:
		actionText = action
	}

	text := fmt.Sprintf("Пользователь %s %s", helpers.UserLink(user), actionText)

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
	text := fmt.Sprintf("Пользователю %s выдано предупреждение (%d/%d)", helpers.UserLink(user), count, maxWarns)

	if until != nil {
		text += fmt.Sprintf(" до %s", helpers.FormatToHumanDateTime(*until))
	}

	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	if banned {
		text += fmt.Sprintf("\n\nПользователь %s за превышение лимита предупреждений.",
			helpers.Gendered(user.Gender, "забанен", "забанена"),
		)
	}

	return text
}

func FormatUnwarnInfo(user model.User, count, maxWarns int) string {
	return fmt.Sprintf("С пользователя %s снято предупреждение (%d/%d)", helpers.UserLink(user), count, maxWarns)
}

func FormatWarnsCleared(user model.User) string {
	return fmt.Sprintf("Все предупреждения пользователя %s были аннулированы", helpers.UserLink(user))
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
		if _, ok := userWarns[w.User.ID]; !ok {
			userOrder = append(userOrder, w.User.ID)
		}
		userWarns[w.User.ID] = append(userWarns[w.User.ID], w)
	}

	for _, userID := range userOrder {
		ws := userWarns[userID]
		u := ws[0].User
		sb.WriteString(fmt.Sprintf("\n👤 %s (активные: %d/%d):\n", helpers.UserLink(u), len(ws), maxWarns))
		for i, w := range ws {
			createdStr := helpers.FormatToHumanDateTime(w.CreatedAt)
			expireStr := ""
			if !w.ExpiresAt.IsZero() {
				expireStr = fmt.Sprintf(", истекает %s", helpers.FormatToHumanDateTime(w.ExpiresAt))
			}
			modName := helpers.UserLink(w.Moderator)
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
		helpers.UserLink(user),
		helpers.Gendered(user.Gender, "размучен", "размучена"),
	)
}

func FormatDirectModerationAction(user model.User, chatTitle string, action string, until *time.Time, reason string) string {
	var actionText string
	switch action {
	case "ban":
		actionText = helpers.Gendered(user.Gender, "забанены", "забанены")
	case "kick":
		actionText = helpers.Gendered(user.Gender, "кикнуты", "кикнуты")
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
