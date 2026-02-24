package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
	"time"
)

func FormatReport(report []model.MessageReportMember, restMembers []model.RestMember, from, to *time.Time) string {
	var periodHeader string
	now := time.Now()

	if from != nil && to != nil {
		periodHeader = fmt.Sprintf("📊 Отчёт за период: %s — %s",
			helpers.FormatToHumanDate(*from),
			helpers.FormatToHumanDate(*to),
		)
	} else if from != nil {
		periodHeader = fmt.Sprintf("📊 Отчёт с %s", helpers.FormatToHumanDate(*from))
	} else if to != nil {
		periodHeader = fmt.Sprintf("📊 Отчёт до %s", helpers.FormatToHumanDate(*to))
	} else {
		periodHeader = fmt.Sprintf("📊 Отчёт за всё время")
	}

	var passed, failedWarn, failedBan, newbies, inRest []string

	for _, r := range report {
		normWarnDone := r.MessagesCount >= r.NormWarn
		normBanDone := true
		if r.NormBan > 0 {
			normBanDone = r.MessagesCount >= r.NormBan
		}
		userTitle := r.CustomTitle
		if r.CustomTitle == "" {
			userTitle = r.User.FirstName
		}

		line := fmt.Sprintf("%s — %d",
			helpers.LinkWithContent(r.User, userTitle),
			r.MessagesCount,
		)

		isNewbie := false
		if r.NewbieThresholdDays > 0 {
			newbieUntil := r.JoinedAt.AddDate(0, 0, int(r.NewbieThresholdDays))
			if newbieUntil.After(now) {
				isNewbie = true
			}
		}

		if isNewbie && normWarnDone {
			line = fmt.Sprintf("%s 🐣 — %d",
				helpers.LinkWithContent(r.User, userTitle),
				r.MessagesCount,
			)
			passed = append(passed, line)
			continue
		}

		if isNewbie {
			newbies = append(newbies, line)
			continue
		}

		if !normBanDone && r.NormBan > 0 {
			failedBan = append(failedBan, line)
		} else if !normWarnDone {
			failedWarn = append(failedWarn, line)
		} else {
			passed = append(passed, line)
		}
	}

	for _, r := range restMembers {
		var untilText string
		if !r.RestUntil.IsZero() {
			untilText = helpers.FormatToHumanDate(r.RestUntil)
		} else {
			untilText = "неизвестно"
		}
		userTitle := r.CustomTitle
		if r.CustomTitle == "" {
			userTitle = r.User.FirstName
		}
		line := fmt.Sprintf("%s до %s",
			helpers.LinkWithContent(r.User, userTitle),
			untilText,
		)
		inRest = append(inRest, line)
	}

	var totalMessages int32
	for _, r := range report {
		totalMessages += r.MessagesCount
	}

	var sb strings.Builder
	sb.WriteString(periodHeader + "\n\n")
	sb.WriteString("<blockquote expandable>")

	sb.WriteString("🌟 Прошли норму\n")
	if len(passed) > 0 {
		writeNumberedList(&sb, passed)
	} else {
		sb.WriteString("Пока никто не прошёл норму\n")
	}

	sb.WriteString("\n❌ Не прошли норму️ (варн) \n")
	if len(failedWarn) > 0 {
		writeNumberedList(&sb, failedWarn)
	} else {
		sb.WriteString("Список пуст\n")
	}

	sb.WriteString("\n🚫 Не прошли норму (бан) \n")
	if len(failedBan) > 0 {
		writeNumberedList(&sb, failedBan)
	} else {
		sb.WriteString("Список пуст\n")
	}

	sb.WriteString("\n🐣 Новички\n")
	if len(newbies) > 0 {
		writeNumberedList(&sb, newbies)
	} else {
		sb.WriteString("Новичков нет\n")
	}

	sb.WriteString("\n💤 Рест\n")
	if len(inRest) > 0 {
		writeNumberedList(&sb, inRest)
	} else {
		sb.WriteString("Пока никто не находится в ресте\n")
	}

	sb.WriteString("</blockquote>")
	sb.WriteString(fmt.Sprintf("\n📝 Всего сообщений: %d\n", totalMessages))

	return sb.String()
}

func writeNumberedList(sb *strings.Builder, items []string) {
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
}
