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

func FormatDevelopersList(users []model.User, roles []string) string {
	var sb strings.Builder
	sb.WriteString("🛠 Разработчики бота:\n")
	for i, u := range users {
		sb.WriteString(fmt.Sprintf("\n%d. %s (%s)", i+1, helpers.Link(u), roles[i]))
	}
	return sb.String()
}

func FormatDeveloperAdded(user model.User, role string) string {
	return fmt.Sprintf("Участник %s %s разработчиком бота с ролью %s",
		helpers.Link(user),
		helpers.Gendered(user.Gender, "назначен", "назначена", "назначен(а)"),
		role,
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
			text += fmt.Sprintf(" до %s", helpers.FormatToHumanDate(*until))
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
		text += fmt.Sprintf(" до %s", helpers.FormatToHumanDate(*until))
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
			text += fmt.Sprintf(" до %s", helpers.FormatToHumanDate(*until))
		} else {
			text += " навсегда"
		}
	}

	if reason != "" {
		text += fmt.Sprintf("\nПричина: %s", reason)
	}

	return text
}
