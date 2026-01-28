package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/exempt"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/stats"
	"fmt"
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

func (h *Handler) ShowStats(b *gotgbot.Bot, ctx *ext.Context, cctx *command.Context) error {
	if _, err := h.memberService.SyncChatMembers(ctx.EffectiveChat.Id); err != nil {
		slog.Warn("failed to auto-update chat members in stats", "chat_id", ctx.EffectiveChat.Id, "error", err)
	}

	var period string
	if len(cctx.Args) == 0 {
		period = "неделя"
	} else {
		period = cctx.Args[0]
	}

	var from, to *time.Time
	switch period {
	case "неделя":
		from, to = stats.ResolvePeriod(stats.PeriodWeek, time.Now())
	case "месяц":
		from, to = stats.ResolvePeriod(stats.PeriodMonth, time.Now())
	case "всё", "все", "всего", "вся":
		from, to = nil, nil
	default:
		from, to = stats.ResolvePeriod(stats.PeriodWeek, time.Now())
	}

	report, err := h.service.GetMemberStats(ctx.EffectiveChat.Id, from, to)
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

	_, err = ctx.EffectiveMessage.Reply(b, formatReport(report, exemptMembers, from, to), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func formatReport(report []model.MessageReportMember, exemptMembers []model.ExemptMember, from, to *time.Time) string {
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

	var passed, failed, newbies, rest []string

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
		line := fmt.Sprintf("%s до %s",
			helpers.LinkWithContent(r.User, fmt.Sprintf("%s (%s)", r.User.FirstName, r.CustomTitle)),
			untilText,
		)
		rest = append(rest, line)
	}

	var totalMessages int32
	for _, r := range report {
		totalMessages += r.MessagesCount
	}

	var sb strings.Builder
	sb.WriteString(periodHeader + "\n\n")

	sb.WriteString("🌟 Прошли норму\n")
	if len(passed) > 0 {
		writeNumberedList(&sb, passed)
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
