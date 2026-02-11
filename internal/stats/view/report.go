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

	var passed, failed, newbies, inRest []string

	for _, r := range report {
		line := fmt.Sprintf("%s — %d сообщений",
			helpers.LinkWithContent(r.User, fmt.Sprintf("%s (%s)", r.User.FirstName, r.CustomTitle)),
			r.MessagesCount,
		)

		isNewbie := false
		if r.NewbieThresholdDays > 0 {
			newbieUntil := r.JoinedAt.AddDate(0, 0, int(r.NewbieThresholdDays))
			if newbieUntil.After(now) {
				isNewbie = true
			}
		}

		if isNewbie && r.NormDone {
			line = fmt.Sprintf("%s 🐣 — %d сообщений",
				helpers.LinkWithContent(r.User, fmt.Sprintf("%s (%s)", r.User.FirstName, r.CustomTitle)),
				r.MessagesCount,
			)
			passed = append(passed, line)
			continue
		}

		if isNewbie {
			newbies = append(newbies, line)
			continue
		}

		if r.NormDone {
			passed = append(passed, line)
		} else {
			failed = append(failed, line)
		}
	}

	for _, r := range restMembers {
		var untilText string
		if !r.RestUntil.IsZero() {
			untilText = helpers.FormatToHumanDate(r.RestUntil)
		} else {
			untilText = "неизвестно"
		}
		line := fmt.Sprintf("%s до %s",
			helpers.LinkWithContent(r.User, fmt.Sprintf("%s (%s)", r.User.FirstName, r.CustomTitle)),
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

	sb.WriteString("🌟 Прошли норму\n")
	if len(passed) > 0 {
		writePassedList(&sb, passed)
	} else {
		sb.WriteString("—\n")
	}

	sb.WriteString("\n❌ Не прошли норму️ \n")
	if len(failed) > 0 {
		writeNumberedList(&sb, failed)
	} else {
		sb.WriteString("—\n")
	}

	sb.WriteString("\n🐣 Новички\n")
	if len(newbies) > 0 {
		writeNumberedList(&sb, newbies)
	} else {
		sb.WriteString("—\n")
	}

	sb.WriteString("\n💤 Рест\n")
	if len(inRest) > 0 {
		writeNumberedList(&sb, inRest)
	} else {
		sb.WriteString("—\n")
	}

	sb.WriteString(fmt.Sprintf("\n📝 Всего сообщений: %d\n", totalMessages))

	return sb.String()
}

func writeNumberedList(sb *strings.Builder, items []string) {
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
}

func writePassedList(sb *strings.Builder, items []string) {
	for i, item := range items {
		prefix := fmt.Sprintf("%d.", i+1)

		switch i {
		case 0:
			prefix = "🔥"
		case 1, 2:
			prefix = "⚡"
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", prefix, item))
	}
}
