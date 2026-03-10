package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
	"time"
)

func FormatReport(report []model.MessageReportMember, restMembers []model.RestMember, from, to *time.Time) string {
	header := formatPeriodHeader(from, to)
	sections := prepareReportSections(report, restMembers)

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	sb.WriteString("<blockquote expandable>")

	sb.WriteString(fmt.Sprintf("%s Прошли норму %d\n", helpers.CustomEmoji(5260446287391630603, "🌟"), sections.NormWarn))
	if len(sections.Passed) > 0 {
		writeNumberedList(&sb, sections.Passed)
	} else {
		sb.WriteString("Список пуст\n")
	}

	sb.WriteString(fmt.Sprintf("\n⚠️ Не прошли норму️ %d (варн) \n", sections.NormWarn))
	if len(sections.FailedWarn) > 0 {
		writeNumberedList(&sb, sections.FailedWarn)
	} else {
		sb.WriteString("Список пуст\n")
	}

	if sections.NormBan != 0 {
		sb.WriteString(fmt.Sprintf("\n🚫 Не прошли норму %d (бан) \n", sections.NormBan))
		if len(sections.FailedBan) > 0 {
			writeNumberedList(&sb, sections.FailedBan)
		} else {
			sb.WriteString("Список пуст\n")
		}
	}

	sb.WriteString("\n🐣 Новички\n")
	if len(sections.Newbies) > 0 {
		writeNumberedList(&sb, sections.Newbies)
	} else {
		sb.WriteString("Список пуст\n")
	}

	sb.WriteString("\n💤 Рест\n")
	if len(sections.InRest) > 0 {
		writeNumberedList(&sb, sections.InRest)
	} else {
		sb.WriteString("Список пуст\n")
	}

	sb.WriteString("</blockquote>")
	sb.WriteString(fmt.Sprintf("\n📝 Всего сообщений: %d\n", sections.TotalMessages))

	return sb.String()
}

func FormatRestList(restMembers []model.RestMember) string {
	if len(restMembers) == 0 {
		return "💤 <b>В ресте никого нет.</b>"
	}

	var inRest []string
	for _, r := range restMembers {
		inRest = append(inRest, formatRestLine(r))
	}

	var sb strings.Builder
	sb.WriteString("💤 <b>Список участников в ресте:</b>\n\n")
	writeNumberedList(&sb, inRest)
	return sb.String()
}

func FormatNewbies(report []model.MessageReportMember) string {
	sections := prepareReportSections(report, nil)

	if len(sections.Newbies) == 0 {
		return "🐣 <b>Новых участников за этот период не найдено.</b>"
	}

	var sb strings.Builder
	sb.WriteString("🐣 <b>Новые участники:</b>\n\n")
	writeNumberedList(&sb, sections.Newbies)
	return sb.String()
}

func FormatFailedNorm(report []model.MessageReportMember, from, to *time.Time) string {
	header := formatPeriodHeader(from, to)
	sections := prepareReportSections(report, nil)

	if len(sections.FailedWarn) == 0 && len(sections.FailedBan) == 0 {
		return header + fmt.Sprintf("\n\n%s <b>Все участники выполнили норму!</b>", helpers.SuccessEmoji())
	}

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	sb.WriteString("⚠️ <b>Не выполнили норму:</b>\n")

	if len(sections.FailedWarn) > 0 {
		sb.WriteString(fmt.Sprintf("\n📉 Меньше %d сообщений (варн):\n", sections.NormWarn))
		writeNumberedList(&sb, sections.FailedWarn)
	}

	if len(sections.FailedBan) > 0 {
		sb.WriteString(fmt.Sprintf("\n🚫 Меньше %d сообщений (бан):\n", sections.NormBan))
		writeNumberedList(&sb, sections.FailedBan)
	}

	return sb.String()
}

type reportSections struct {
	Passed        []string
	FailedWarn    []string
	FailedBan     []string
	Newbies       []string
	InRest        []string
	NormWarn      int32
	NormBan       int32
	TotalMessages int32
}

func prepareReportSections(report []model.MessageReportMember, restMembers []model.RestMember) reportSections {
	now := time.Now().In(helpers.MoscowLocation)
	s := reportSections{}

	for _, r := range report {
		s.NormBan = r.NormBan
		s.NormWarn = r.NormWarn
		s.TotalMessages += r.MessagesCount

		normWarnDone := r.MessagesCount >= r.NormWarn
		normBanDone := true
		if r.NormBan > 0 {
			normBanDone = r.MessagesCount >= r.NormBan
		}

		userTitle := r.CustomTitle
		if r.CustomTitle == "" {
			userTitle = r.User.FirstName
		}

		line := fmt.Sprintf("%s — %d", helpers.LinkWithContent(r.User, userTitle), r.MessagesCount)

		isNewbie := false
		if r.NewbieThresholdDays > 0 {
			newbieUntil := r.JoinedAt.AddDate(0, 0, int(r.NewbieThresholdDays))
			if newbieUntil.After(now) {
				isNewbie = true
			}
		}

		if isNewbie {
			if normWarnDone {
				s.Passed = append(s.Passed, fmt.Sprintf("%s 🐣 — %d", helpers.LinkWithContent(r.User, userTitle), r.MessagesCount))
			} else {
				s.Newbies = append(s.Newbies, line)
			}
			continue
		}

		if !normBanDone && r.NormBan > 0 {
			s.FailedBan = append(s.FailedBan, line)
		} else if !normWarnDone {
			s.FailedWarn = append(s.FailedWarn, line)
		} else {
			s.Passed = append(s.Passed, line)
		}
	}

	for _, r := range restMembers {
		s.InRest = append(s.InRest, formatRestLine(r))
	}

	return s
}

func formatRestLine(r model.RestMember) string {
	var untilText string
	if !r.RestUntil.IsZero() {
		untilText = helpers.FormatToHumanDateTime(r.RestUntil)
	} else {
		untilText = "неизвестно"
	}
	userTitle := r.CustomTitle
	if r.CustomTitle == "" {
		userTitle = r.User.FirstName
	}
	return fmt.Sprintf("%s до %s", helpers.LinkWithContent(r.User, userTitle), untilText)
}

func formatPeriodHeader(from, to *time.Time) string {
	if from != nil && to != nil {
		return fmt.Sprintf("📊 Отчёт за период: %s — %s", helpers.FormatToHumanDateTime(*from), helpers.FormatToHumanDateTime(*to))
	} else if from != nil {
		return fmt.Sprintf("📊 Отчёт с %s", helpers.FormatToHumanDateTime(*from))
	} else if to != nil {
		return fmt.Sprintf("📊 Отчёт до %s", helpers.FormatToHumanDateTime(*to))
	}
	return "📊 Отчёт за всё время"
}

func writeNumberedList(sb *strings.Builder, items []string) {
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
}
