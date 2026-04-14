package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"

	"github.com/gotd/td/telegram/message/entity"
)

func StatusTitle(status model.Status, count int) string {
	if count == 1 {
		return status.Title()
	}

	return status.PluralTitle()
}
func FormatAdminsList(admins []model.ChatMember) string {
	eb := &entity.Builder{}
	WriteAdminsList(eb, admins)
	res, _ := eb.Complete()
	return res
}

func WriteAdminsList(eb *entity.Builder, admins []model.ChatMember) {
	categories := map[model.Status][]model.ChatMember{}

	for _, a := range admins {
		if !a.LeftAt.IsZero() {
			continue
		}
		categories[a.Status] = append(categories[a.Status], a)
	}

	order := [5]model.Status{model.StatusOwner, model.StatusCoOwner, model.StatusSeniorAdmin, model.StatusAdmin, model.StatusModerator}
	for _, status := range order {
		members := categories[status]
		if len(members) == 0 {
			continue
		}

		eb.Plain("\n")
		helpers.WriteCustomEmoji(eb, helpers.StatusEmojiID(status), helpers.StatusEmojiPlain(status))
		eb.Plain(" " + StatusTitle(status, len(members)) + "\n")

		for _, m := range members {
			eb.Plain("▸ ")
			helpers.WriteRoleEmojiLink(eb, m)
			eb.Plain("\n")
		}
	}

	eb.Plain("\nЧтобы добавить, напишите: ")
	eb.Code("фм повысить @участник ранг")
}

func WriteAdminAdded(eb *entity.Builder, user model.ChatMember, status model.Status) {
	eb.Plain("Участнику ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" ")
	eb.Plain(helpers.Gendered(user.User.Gender, "назначен", "назначена"))
	eb.Plain(fmt.Sprintf(" ранг %d (%s)", status, status.String()))
}

func WriteAdminRemoved(eb *entity.Builder, user model.ChatMember) {
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" ")
	eb.Plain(helpers.Gendered(user.User.Gender, "удален", "удалена"))
	eb.Plain(" из администраторов бота")
}

func FormatAdminAdded(user model.ChatMember, status model.Status) string {
	eb := &entity.Builder{}
	WriteAdminAdded(eb, user, status)
	res, _ := eb.Complete()
	return res
}

func WriteModerationAction(eb *entity.Builder, user model.ChatMember, action string, until time.Time, reason string) {
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

	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" ")
	eb.Plain(actionText)

	if action != "kick" {
		if !until.IsZero() {
			eb.Plain(" до ")
			helpers.FormattedDate(eb, until)
		} else {
			eb.Plain(" навсегда")
		}
	}

	if reason != "" {
		eb.Plain("\nПричина: ")
		eb.Plain(reason)
	}
}

func FormatModerationAction(user model.ChatMember, action string, until time.Time, reason string) string {
	eb := &entity.Builder{}
	WriteModerationAction(eb, user, action, until, reason)
	res, _ := eb.Complete()
	return res
}

func WriteWarnInfo(eb *entity.Builder, user model.ChatMember, count, maxWarns int, until time.Time, reason string, banned bool) {
	eb.Plain("Участнику ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(fmt.Sprintf(" выдано предупреждение (%d/%d)", count, maxWarns))

	if !until.IsZero() {
		eb.Plain(" до ")
		helpers.FormattedDate(eb, until)
	}

	if reason != "" {
		eb.Plain("\nПричина: ")
		eb.Plain(reason)
	}

	if banned {
		eb.Plain("\n\nУчастник ")
		eb.Plain(helpers.Gendered(user.User.Gender, "забанен", "забанена"))
		eb.Plain(" за превышение лимита предупреждений.")
	}
}

func FormatWarnInfo(user model.ChatMember, count, maxWarns int, until time.Time, reason string, banned bool) string {
	eb := &entity.Builder{}
	WriteWarnInfo(eb, user, count, maxWarns, until, reason, banned)
	res, _ := eb.Complete()
	return res
}

func FormatUnwarnInfo(user model.ChatMember, count, maxWarns int) string {
	eb := &entity.Builder{}
	WriteUnwarnInfo(eb, user, count, maxWarns)
	res, _ := eb.Complete()
	return res
}

func FormatWarnsCleared(user model.ChatMember) string {
	eb := &entity.Builder{}
	WriteWarnsCleared(eb, user)
	res, _ := eb.Complete()
	return res
}

func WriteWarnsCleared(eb *entity.Builder, user model.ChatMember) {
	eb.Plain("Все предупреждения участника ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" были аннулированы")
}

func WriteWarnlist(eb *entity.Builder, warns []model.Warn, maxWarns int) {
	if len(warns) == 0 {
		helpers.WriteSuccessEmoji(eb)
		eb.Plain(" В этом чате нет активных предупреждений")
		return
	}

	eb.Bold("⚠️ Список всех предупреждений в чате:")
	eb.Plain("\n")

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
		eb.Plain("\n👤 ")
		helpers.WriteRoleEmojiLink(eb, ws[0].ChatMember)
		eb.Plain(fmt.Sprintf(" (активные: %d/%d):\n", len(ws), maxWarns))

		for i, w := range ws {
			eb.Plain(fmt.Sprintf("  %d. Выдан ", i+1))
			helpers.FormattedDate(eb, w.CreatedAt)
			eb.Plain(" модератором ")
			helpers.WriteRoleEmojiLink(eb, w.Moderator)
			if !w.ExpiresAt.IsZero() {
				eb.Plain(", истекает ")
				helpers.FormattedDate(eb, w.ExpiresAt)
			}
			if w.Reason != "" {
				eb.Plain(fmt.Sprintf(", причина: %s", w.Reason))
			}
			eb.Plain("\n")
		}
	}
}

func FormatWarnlist(warns []model.Warn, maxWarns int) string {
	eb := &entity.Builder{}
	WriteWarnlist(eb, warns, maxWarns)
	res, _ := eb.Complete()
	return res
}

func FormatUnmuteInfo(user model.ChatMember) string {
	eb := &entity.Builder{}
	WriteUnmuteInfo(eb, user)
	res, _ := eb.Complete()
	return res
}

func WriteUnmuteInfo(eb *entity.Builder, user model.ChatMember) {
	eb.Plain("Участник ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" ")
	eb.Plain(helpers.Gendered(user.User.Gender, "размучен", "размучена"))
}

func WriteUnwarnInfo(eb *entity.Builder, user model.ChatMember, count int, max int) {
	eb.Plain("Предупреждение участнику ")
	helpers.WriteRoleEmojiLink(eb, user)
	eb.Plain(" ")
	eb.Plain(helpers.Gendered(user.User.Gender, "снято", "снято"))
	eb.Plain(fmt.Sprintf(" (активные: %d/%d)", count, max))
}

func WriteDirectModerationAction(eb *entity.Builder, user model.ChatMember, chatTitle string, action string, until time.Time, reason string) {
	var actionText string
	switch action {
	case "ban":
		actionText = helpers.Gendered(user.User.Gender, "забанены", "забанены")
	case "kick":
		actionText = helpers.Gendered(user.User.Gender, "кикнуты", "кикнуты")
	default:
		actionText = action
	}

	eb.Plain("Вы были ")
	eb.Plain(actionText)
	eb.Plain(" в чате ")
	eb.Bold(chatTitle)

	if action == "ban" {
		if !until.IsZero() {
			eb.Plain(" до ")
			helpers.FormattedDate(eb, until)
		} else {
			eb.Plain(" навсегда")
		}
	}

	if reason != "" {
		eb.Plain("\nПричина: ")
		eb.Plain(reason)
	}
}

func FormatDirectModerationAction(user model.ChatMember, chatTitle string, action string, until time.Time, reason string) string {
	eb := &entity.Builder{}
	WriteDirectModerationAction(eb, user, chatTitle, action, until, reason)
	res, _ := eb.Complete()
	return res
}
