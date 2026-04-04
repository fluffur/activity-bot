package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"

	"github.com/gotd/td/telegram/message/entity"
)

func WriteStats(eb *entity.Builder, report []model.ChatMemberMessageCount, restMembers []model.ChatMember, newbieThresholdDays int32, from, to *time.Time) {
	WritePeriodHeader(eb, from, to)
	eb.Plain("\n\n")

	sections := prepareReportSections(report, restMembers)
	topOnly := sections.NormWarn == 0 && sections.NormBan == 0

	if !topOnly {
		if sections.NormWarn > 0 {
			helpers.WriteCustomEmoji(eb, "5870633910337015697", "✅")
			eb.Plain(fmt.Sprintf(" Прошли норму %d\n", sections.NormWarn))

			token := eb.Token()
			if len(sections.Passed) > 0 {
				writeNumberedList(eb, sections.Passed)
			} else {
				eb.Plain("Список пуст\n")
			}
			token.Apply(eb, entity.Blockquote(true))

			eb.Plain("\n")
			helpers.WriteCustomEmoji(eb, "5870948572526022116", "⚠️")
			eb.Plain(fmt.Sprintf(" Не прошли норму️ %d (варн) \n", sections.NormWarn))

			token = eb.Token()
			if len(sections.FailedWarn) > 0 {
				writeNumberedList(eb, sections.FailedWarn)
			} else {
				eb.Plain("Список пуст\n")
			}
			token.Apply(eb, entity.Blockquote(true))
		}

		if sections.NormBan > 0 {
			eb.Plain("\n")
			helpers.WriteCustomEmoji(eb, "5870657884844462243", "❌")
			eb.Plain(fmt.Sprintf(" Не прошли норму %d (бан) \n", sections.NormBan))

			token := eb.Token()
			if len(sections.FailedBan) > 0 {
				writeNumberedList(eb, sections.FailedBan)
			} else {
				eb.Plain("Список пуст\n")
			}
			token.Apply(eb, entity.Blockquote(true))
		}

		if newbieThresholdDays > 0 {
			eb.Plain("\n")
			helpers.WriteNewbieEmoji(eb)
			eb.Plain(fmt.Sprintf(" Новички (%d %s)\n", newbieThresholdDays, helpers.PluralizeDays(int(newbieThresholdDays))))

			token := eb.Token()
			if len(sections.Newbies) > 0 {
				writeNumberedList(eb, sections.Newbies)
			} else {
				eb.Plain("Список пуст\n")
			}
			token.Apply(eb, entity.Blockquote(true))
		}

		eb.Plain("\n")
		helpers.WriteRestEmoji(eb)
		eb.Plain(" Рест\n")

		token := eb.Token()
		if len(sections.InRest) > 0 {
			for i, r := range sections.InRest {
				eb.Plain(fmt.Sprintf("%d. ", i+1))
				writeRestLine(eb, r)
				eb.Plain("\n")
			}
		} else {
			eb.Plain("Список пуст\n")
		}
		token.Apply(eb, entity.Blockquote(true))

	} else {
		eb.CustomEmoji("🌟", 5224694451338759997)
		eb.Plain(" Топ участников\n")

		token := eb.Token()
		if len(sections.Passed) > 0 {
			writeNumberedList(eb, sections.Passed)
		} else {
			eb.Plain("Список пуст\n")
		}
		token.Apply(eb, entity.Blockquote(true))
	}

	eb.Plain("\n")
	helpers.WriteTotalEmoji(eb)
	eb.Plain(" Всего сообщений: ")
	eb.Code(fmt.Sprintf("%d", sections.TotalMessages))
	eb.Plain("\n")
}

func WriteRestList(eb *entity.Builder, restMembers []model.ChatMember) {
	if len(restMembers) == 0 {
		helpers.WriteRestEmoji(eb)
		eb.Plain(" ")
		eb.Bold("В ресте никого нет.")
		return
	}

	helpers.WriteRestEmoji(eb)
	eb.Plain(" ")
	eb.Bold("Список участников в ресте:")
	eb.Plain("\n\n")

	for i, r := range restMembers {
		eb.Plain(fmt.Sprintf("%d. ", i+1))
		writeRestLine(eb, r)
		eb.Plain("\n")
	}
}

func WriteNewbies(eb *entity.Builder, report []model.ChatMemberMessageCount) {
	sections := prepareReportSections(report, nil)

	if len(sections.Newbies) == 0 {
		helpers.WriteNewbieEmoji(eb)
		eb.Plain(" ")
		eb.Bold("Новых участников за этот период не найдено.")
		return
	}

	helpers.WriteNewbieEmoji(eb)
	eb.Plain(" ")
	eb.Bold("Новые участники:")
	eb.Plain("\n\n")

	writeNumberedList(eb, sections.Newbies)
}

func WriteFailedNorm(eb *entity.Builder, report []model.ChatMemberMessageCount, from, to *time.Time) {
	WritePeriodHeader(eb, from, to)
	sections := prepareReportSections(report, nil)

	if len(sections.FailedWarn) == 0 && len(sections.FailedBan) == 0 {
		eb.Plain("\n\n")
		helpers.WriteSuccessEmoji(eb)
		eb.Plain(" ")
		eb.Bold("Все участники выполнили норму!")
		return
	}

	eb.Plain("\n\n⚠️ ")
	eb.Bold("Не выполнили норму:")
	eb.Plain("\n")

	if len(sections.FailedWarn) > 0 {
		eb.Plain("\n")
		helpers.WriteCustomEmoji(eb, "5224340348465073584", "⚠️")
		eb.Plain(fmt.Sprintf(" Меньше %d сообщений (варн):\n", sections.NormWarn))
		writeNumberedList(eb, sections.FailedWarn)
	}

	if len(sections.FailedBan) > 0 {
		eb.Plain("\n")
		helpers.WriteDangerEmoji(eb)
		eb.Plain(fmt.Sprintf(" Меньше %d сообщений (бан):\n", sections.NormBan))
		writeNumberedList(eb, sections.FailedBan)
	}
}

type reportSections struct {
	Passed        []model.ChatMemberMessageCount
	FailedWarn    []model.ChatMemberMessageCount
	FailedBan     []model.ChatMemberMessageCount
	Newbies       []model.ChatMemberMessageCount
	InRest        []model.ChatMember
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

		isNewbie := false
		if r.Chat.NewbieThresholdDays > 0 {
			newbieUntil := r.ChatMember.JoinedAt.AddDate(0, 0, int(r.Chat.NewbieThresholdDays))
			if newbieUntil.After(now) {
				isNewbie = true
			}
		}

		if isNewbie {
			if normWarnDone {
				s.Passed = append(s.Passed, r)
			} else {
				s.Newbies = append(s.Newbies, r)
			}
			continue
		}

		if !normBanDone && r.Chat.NormBan > 0 {
			s.FailedBan = append(s.FailedBan, r)
		} else if !normWarnDone {
			s.FailedWarn = append(s.FailedWarn, r)
		} else {
			s.Passed = append(s.Passed, r)
		}
	}

	s.InRest = restMembers

	return s
}

func writeRestLine(eb *entity.Builder, r model.ChatMember) {
	var untilText string
	if !r.RestUntil.IsZero() {
		untilText = helpers.FormatToHumanDateTime(r.RestUntil)
	} else {
		untilText = "неизвестно"
	}
	helpers.WriteRoleEmojiLink(eb, r)
	eb.Plain(fmt.Sprintf(" до %s", untilText))
}

func WritePeriodHeader(eb *entity.Builder, from, to *time.Time) {
	helpers.WriteCustomEmoji(eb, "5870772616305839506", "📊")
	if from != nil && to != nil {
		eb.Plain(" Отчет за период:\n")
		helpers.FormattedDate(eb, *from)
		eb.Plain("\n")
		helpers.FormattedDate(eb, *to)

	} else if from != nil {
		eb.Plain(fmt.Sprintf(" Отчет с %s", helpers.FormatToHumanDateTime(*from)))
	} else if to != nil {
		eb.Plain(fmt.Sprintf(" Отчет до %s", helpers.FormatToHumanDateTime(*to)))
	} else {
		eb.Plain(" Отчет за всё время")
	}
}

func writeNumberedList(eb *entity.Builder, items []model.ChatMemberMessageCount) {
	for i, item := range items {
		eb.Plain(fmt.Sprintf("%d. ", i+1))
		helpers.WriteRoleEmojiLink(eb, item.ChatMember)
		eb.Plain(" — ")
		eb.Code(fmt.Sprintf("%d", item.MessageCount))
		eb.Plain("\n")
	}
}
