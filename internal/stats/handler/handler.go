package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/exempt"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/stats"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service       *stats.Service
	exemptService *exempt.Service
	memberService *member.Service
}

func New(service *stats.Service, exemptService *exempt.Service, memberService *member.Service) *Handler {
	return &Handler{service, exemptService, memberService}
}

func (h *Handler) ShowStats(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	if _, err := h.memberService.SyncChatMembers(ctx.EffectiveChat.Id); err != nil {
		slog.Warn("failed to auto-update chat members in stats", "chat_id", ctx.EffectiveChat.Id, "error", err)
	}

	report, err := h.service.GetMemberStats(ctx.EffectiveChat.Id)
	if err != nil {
		slog.Error("failed to get member stats for report", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось получить отчёт", nil)
		return err

	}

	exemptMembers, err := h.exemptService.GetExemptMembers(ctx.EffectiveChat.Id)
	if err != nil {
		slog.Error("failed to get exempt members for report", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось получить отчёт", nil)
		return err
	}

	if len(report) == 0 && len(exemptMembers) == 0 {
		_, err = ctx.EffectiveMessage.Reply(b, "Нет данных для отчёта на эту неделю", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, formatWeeklyReport(report, exemptMembers), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}
func formatWeeklyReport(report []model.WeeklyMessageReportMember, exemptMembers []model.ExemptMember) string {
	now := time.Now()

	weekday := int(now.Weekday())
	daysSinceMonday := (weekday + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	sunday := monday.AddDate(0, 0, 6)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())

	weekHeader := fmt.Sprintf("📊 Отчёт за неделю: %s — %s", helpers.FormatToHumanDate(monday), helpers.FormatToHumanDate(sunday))

	var passed, failed, newbies, rest []string

	for _, r := range report {
		line := fmt.Sprintf("%s — %d сообщений", helpers.Link(r.User), r.MessagesCount)

		isNewbie := false
		if r.NewbieThresholdDays > 0 {
			newbieUntil := r.JoinedAt.AddDate(0, 0, int(r.NewbieThresholdDays))
			if newbieUntil.After(now) {
				log.Println(newbieUntil.String())
				isNewbie = true
			}
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

	for _, r := range exemptMembers {
		var untilText string
		if !r.ExemptUntil.IsZero() {
			untilText = helpers.FormatToHumanDate(r.ExemptUntil)
		} else {
			untilText = "неизвестно"
		}
		line := fmt.Sprintf("%s до %s", helpers.Link(r.User), untilText)
		rest = append(rest, line)
	}
	var totalMessages int32 = 0
	for _, r := range report {
		totalMessages += r.MessagesCount
	}

	var sb strings.Builder
	sb.WriteString(weekHeader + "\n\n")

	sb.WriteString("✅ Прошли норму 🌟\n")
	if len(passed) > 0 {
		writeNumberedList(&sb, passed)
	} else {
		sb.WriteString("—\n")
	}

	sb.WriteString("\n❌ Не прошли норму ⚠️ \n")
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
	if len(rest) > 0 {
		writeNumberedList(&sb, rest)
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
