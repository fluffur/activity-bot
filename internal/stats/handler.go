package stats

import (
	"activity-bot/internal/chat/member"
	"activity-bot/internal/command"
	"activity-bot/internal/exempt"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service       *Service
	exemptService *exempt.Service
	memberService *member.Service
}

func NewHandler(service *Service, exemptService *exempt.Service, memberService *member.Service) *Handler {
	return &Handler{service, exemptService, memberService}
}

func (h *Handler) ShowStats(b *gotgbot.Bot, ctx *ext.Context, _ *command.Context) error {
	if _, err := member.UpdateChatMembers(b, h.memberService, ctx.EffectiveChat.Id); err != nil {
		log.Println("Auto-update chat members error", err)
	}

	report, err := h.service.GetMemberStats(ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("Exists member stats error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось получить отчёт", nil)
		return err

	}

	exemptMembers, err := h.exemptService.GetExemptMembers(ctx.EffectiveChat.Id)
	if err != nil {
		log.Println("Exists exempt members error", err)
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

	var passed, failed, rest []string

	for _, r := range report {
		line := fmt.Sprintf("%s — %d сообщений", helpers.Link(r.User), r.MessagesCount)
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
