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
	return fmt.Sprintf("Пользователь %s назначен администратором бота", helpers.Link(user))
}

func FormatAdminRemoved(user model.User) string {
	return fmt.Sprintf("Пользователь %s удалён из администраторов бота", helpers.Link(user))
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
	return fmt.Sprintf("Пользователь %s назначен разработчиком бота с ролью %s", helpers.Link(user), role)
}

func FormatDeveloperRemoved(user model.User) string {
	return fmt.Sprintf("Пользователь %s удален из списка разработчиков", helpers.Link(user))
}

func FormatModerationAction(user model.User, action string, until *time.Time, reason string) string {
	var actionText string
	switch action {
	case "ban":
		actionText = "забанен"
	case "mute":
		actionText = "замучен"
	case "kick":
		actionText = "был кикнут из чата"
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
		text += "\n\nПользователь забанен за превышение лимита предупреждений."
	}

	return text
}

func FormatWarnsCount(user model.User, count, maxWarns int) string {
	return fmt.Sprintf("У пользователя %s %d/%d предупреждений", helpers.Link(user), count, maxWarns)
}

func FormatUnwarnInfo(user model.User, count, maxWarns int) string {
	return fmt.Sprintf("С пользователя %s снято предупреждение (%d/%d)", helpers.Link(user), count, maxWarns)
}

func FormatWarnsCleared(user model.User) string {
	return fmt.Sprintf("Все предупреждения пользователя %s были аннулированы", helpers.Link(user))
}

func FormatUnbanInfo(user model.User) string {
	return fmt.Sprintf("Пользователь %s разбанен", helpers.Link(user))
}

func FormatUnmuteInfo(user model.User) string {
	return fmt.Sprintf("Пользователь %s размучен", helpers.Link(user))
}
