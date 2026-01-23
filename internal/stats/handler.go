package stats

import (
	"activity-bot/internal/chat/member"
	"activity-bot/internal/exempt"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	service       *Service
	exemptService *exempt.Service
	memberService *member.Service
}

func NewHandler(service *Service, exemptService *exempt.Service, memberService *member.Service) *Handler {
	return &Handler{service, exemptService, memberService}
}

func (h *Handler) ShowWeeklyReport(ctx context.Context, b *bot.Bot, update *models.Update) {
	if _, err := helpers.UpdateChatMembers(ctx, b, h.memberService, update.Message.Chat.ID); err != nil {
		log.Println("Auto-update chat members error", err)
	}

	report, err := h.service.GetMemberStats(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println("Get member stats error", err)
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить отчёт")
		return
	}

	exemptMembers, err := h.exemptService.GetExemptMembers(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println("Get exempt members error", err)
		helpers.AnswerMessage(ctx, b, update, "Не удалось получить отчёт")
		return
	}

	if len(report) == 0 && len(exemptMembers) == 0 {
		helpers.AnswerMessage(ctx, b, update, "Нет данных для отчёта на эту неделю")
		return
	}

	text := formatWeeklyReport(report, exemptMembers)
	helpers.AnswerMessage(ctx, b, update, text)
}

func formatWeeklyReport(report []model.WeeklyMessageReportMember, exemptMembers []model.ExemptMember) string {
	now := time.Now()

	weekday := int(now.Weekday())
	daysSinceMonday := (weekday + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	sunday := monday.AddDate(0, 0, 6)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())

	weekHeader := fmt.Sprintf("📊 Отчёт за неделю: %s — %s", monday.Format("02.01.2006"), sunday.Format("02.01.2006"))

	var passed, failed, rest []string

	for _, r := range report {
		line := fmt.Sprintf(`%s: %d`, helpers.FormatSilentMentionHTML(r.User), r.MessagesCount)

		if r.NormDone {
			passed = append(passed, line)
		} else {
			failed = append(failed, line)
		}
	}

	for _, r := range exemptMembers {
		var untilText string
		if !r.ExemptUntil.IsZero() {
			untilText = helpers.FormatToHumanDate(r.ExemptUntil)
		} else {
			untilText = "неизвестно"
		}

		line := fmt.Sprintf(`%s до %s`, helpers.FormatSilentMentionHTML(r.User), untilText)
		rest = append(rest, line)
	}

	var sb strings.Builder
	sb.WriteString(weekHeader)

	if len(passed) > 0 {
		sb.WriteString("\n\n✅ Прошли норму\n")
		writeNumberedList(&sb, passed)
		sb.WriteString("\n")
	}
	if len(failed) > 0 {
		sb.WriteString("\n❎ Не прошли норму\n")
		writeNumberedList(&sb, failed)
		sb.WriteString("\n")
	}
	if len(rest) > 0 {
		sb.WriteString("\n💛 Рест\n")
		writeNumberedList(&sb, rest)
	}

	return sb.String()
}

func writeNumberedList(sb *strings.Builder, items []string) {
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
}
