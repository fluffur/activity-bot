package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"
	"time"
)

func FormatStats(report []model.ChatMemberMessageCount, restMembers []model.ChatMember, newbieThresholdDays int32, from, to *time.Time) string {
	header := formatPeriodHeader(from, to)
	sections := prepareReportSections(report, restMembers)
	topOnly := sections.NormWarn == 0 && sections.NormBan == 0

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	if !topOnly {

		if sections.NormWarn > 0 {
			sb.WriteString(fmt.Sprintf("%s Прошли норму %d\n", helpers.CustomEmoji("5870633910337015697", "✅"), sections.NormWarn))

			sb.WriteString("<blockquote expandable>")
			if len(sections.Passed) > 0 {
				writeNumberedList(&sb, sections.Passed)
			} else {
				sb.WriteString("Список пуст\n")
			}
			sb.WriteString("</blockquote>")

			sb.WriteString(fmt.Sprintf("\n%s Не прошли норму️ %d (варн) \n", helpers.CustomEmoji("5870948572526022116", "⚠️"), sections.NormWarn))
			sb.WriteString("<blockquote expandable>")
			if len(sections.FailedWarn) > 0 {
				writeNumberedList(&sb, sections.FailedWarn)
			} else {
				sb.WriteString("Список пуст\n")
			}
			sb.WriteString("</blockquote>")
		}

		if sections.NormBan > 0 {
			sb.WriteString(fmt.Sprintf("\n%s Не прошли норму %d (бан) \n", helpers.CustomEmoji("5870657884844462243", "❌"), sections.NormBan))
			sb.WriteString("<blockquote expandable>")
			if len(sections.FailedBan) > 0 {
				writeNumberedList(&sb, sections.FailedBan)
			} else {
				sb.WriteString("Список пуст\n")
			}
			sb.WriteString("</blockquote>")
		}

		sb.WriteString(fmt.Sprintf("\n%s Новички (%d %s)\n", helpers.NewbieEmoji(), newbieThresholdDays, helpers.PluralizeDays(int(newbieThresholdDays))))
		sb.WriteString("<blockquote expandable>")
		if len(sections.Newbies) > 0 {
			writeNumberedList(&sb, sections.Newbies)
		} else {
			sb.WriteString("Список пуст\n")
		}
		sb.WriteString("</blockquote>")

		sb.WriteString(fmt.Sprintf("\n%s Рест\n", helpers.RestEmoji()))
		sb.WriteString("<blockquote expandable>")
		if len(sections.InRest) > 0 {
			writeNumberedList(&sb, sections.InRest)

		} else {
			sb.WriteString("Список пуст\n")
		}
		sb.WriteString("</blockquote>")

	} else {

		sb.WriteString(fmt.Sprintf("%s Топ участников\n", helpers.CustomEmoji("5224694451338759997", "🌟")))

		sb.WriteString("<blockquote expandable>")
		if len(sections.Passed) > 0 {
			writeNumberedList(&sb, sections.Passed)
		} else {
			sb.WriteString("Список пуст\n")
		}
		sb.WriteString("</blockquote>")
	}

	sb.WriteString(fmt.Sprintf("\n%s Всего сообщений: <code>%d</code>\n", helpers.TotalEmoji(), sections.TotalMessages))

	return sb.String()
}

func FormatRestList(restMembers []model.ChatMember) string {
	if len(restMembers) == 0 {
		return fmt.Sprintf("%s <b>В ресте никого нет.</b>", helpers.RestEmoji())
	}

	var inRest []string
	for _, r := range restMembers {
		inRest = append(inRest, formatRestLine(r))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s <b>Список участников в ресте:</b>\n\n", helpers.RestEmoji()))
	writeNumberedList(&sb, inRest)
	return sb.String()
}

func FormatNewbies(report []model.ChatMemberMessageCount) string {
	sections := prepareReportSections(report, nil)

	if len(sections.Newbies) == 0 {
		return fmt.Sprintf("%s <b>Новых участников за этот период не найдено.</b>", helpers.NewbieEmoji())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s <b>Новые участники:</b>\n\n", helpers.NewbieEmoji()))
	writeNumberedList(&sb, sections.Newbies)
	return sb.String()
}

func FormatFailedNorm(report []model.ChatMemberMessageCount, from, to *time.Time) string {
	header := formatPeriodHeader(from, to)
	sections := prepareReportSections(report, nil)

	if len(sections.FailedWarn) == 0 && len(sections.FailedBan) == 0 {
		return header + fmt.Sprintf("\n\n%s <b>Все участники выполнили норму!</b>", helpers.SuccessEmoji())
	}

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	sb.WriteString("⚠️ <b>Не выполнили норму:</b>\n")

	if len(sections.FailedWarn) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Меньше %d сообщений (варн):\n", helpers.CustomEmoji("5224340348465073584", "⚠️"), sections.NormWarn))
		writeNumberedList(&sb, sections.FailedWarn)
	}

	if len(sections.FailedBan) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Меньше %d сообщений (бан):\n", helpers.DangerEmoji(), sections.NormBan))
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
	TotalMessages int64
}

func prepareReportSections(report []model.ChatMemberMessageCount, restMembers []model.ChatMember) reportSections {
	now := time.Now().UTC()
	s := reportSections{}

	for _, r := range report {
		s.NormBan = r.Chat.NormBan
		s.NormWarn = r.Chat.NormWarn
		s.TotalMessages += r.MessageCount

		normWarnDone := r.MessageCount >= int64(r.Chat.NormWarn)
		normBanDone := true
		if r.Chat.NormBan > 0 {
			normBanDone = r.MessageCount >= int64(r.Chat.NormBan)
		}

		line := fmt.Sprintf("%s — <code>%d</code>", helpers.RoleEmojiLink(r.ChatMember), r.MessageCount)

		isNewbie := false
		if r.Chat.NewbieThresholdDays > 0 {
			newbieUntil := r.ChatMember.JoinedAt.AddDate(0, 0, int(r.Chat.NewbieThresholdDays))
			if newbieUntil.After(now) {
				isNewbie = true
			}
		}

		if isNewbie {
			if normWarnDone {
				s.Passed = append(s.Passed, fmt.Sprintf("%s %s — <code>%d</code>", helpers.NewbieEmoji(), helpers.RoleEmojiLink(r.ChatMember), r.MessageCount))
			} else {
				s.Newbies = append(s.Newbies, line)
			}
			continue
		}

		if !normBanDone && r.Chat.NormBan > 0 {
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

func formatRestLine(r model.ChatMember) string {
	var untilText string
	if !r.RestUntil.IsZero() {
		untilText = helpers.FormatToHumanDateTime(r.RestUntil)
	} else {
		untilText = "неизвестно"
	}
	return fmt.Sprintf("%s до %s", helpers.RoleEmojiLink(r), untilText)
}

func formatPeriodHeader(from, to *time.Time) string {
	if from != nil && to != nil {
		return fmt.Sprintf("%s Отчет за период\n%s — %s", helpers.CustomEmoji("5870772616305839506", "📊"), helpers.FormatToHumanDateTime(*from), helpers.FormatToHumanDateTime(*to))
	} else if from != nil {
		return fmt.Sprintf("%s Отчет с %s", helpers.CustomEmoji("5870772616305839506", "📊"), helpers.FormatToHumanDateTime(*from))
	} else if to != nil {
		return fmt.Sprintf("%s Отчет до %s", helpers.CustomEmoji("5870772616305839506", "📊"), helpers.FormatToHumanDateTime(*to))
	}
	return fmt.Sprintf("%s Отчет за всё время", helpers.CustomEmoji("5870772616305839506", "📊"))
}

func writeNumberedList(sb *strings.Builder, items []string) {
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
}
