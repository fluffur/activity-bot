package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/rest"
	"activity-bot/internal/stats"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service       *stats.Service
	restService   *rest.Service
	memberService *member.Service
}

func New(service *stats.Service, restService *rest.Service, memberService *member.Service) *Handler {
	return &Handler{service, restService, memberService}
}

func (h *Handler) ShowStats(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	if _, err := h.memberService.SyncChatMembers(ctx.EffectiveChat.Id); err != nil {
		slog.Warn("failed to auto-update chat members in stats", "chat_id", ctx.EffectiveChat.Id, "error", err)
	}

	var period string
	if len(cctx.Args()) == 0 {
		period = "неделя"
	} else {
		period = cctx.FirstArgument()
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

	report, err := h.service.GetAllMembersStats(ctx.EffectiveChat.Id, from, to)
	if err != nil {
		slog.Error("failed to get member stats for report", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось получить отчёт", nil)
		return err

	}

	restMembers, err := h.restService.GetRestMembers(ctx.EffectiveChat.Id)
	if err != nil {
		slog.Error("failed to get rest members for report", "chat_id", ctx.EffectiveChat.Id, "error", err)
		_, err = ctx.EffectiveMessage.Reply(b, "Не удалось получить отчёт", nil)
		return err
	}

	if len(report) == 0 && len(restMembers) == 0 {
		_, err = ctx.EffectiveMessage.Reply(b, "Нет данных для отчёта на эту неделю", nil)
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(b, formatReport(report, restMembers, from, to), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
	return err
}

func (h *Handler) WhoAmI(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	return h.WhoAreUser(b, ctx, ctx.EffectiveSender.Id())

}

func (h *Handler) WhoAreYou(b *gotgbot.Bot, ctx *ext.Context, cctx *cmd.Context) error {
	u := cctx.FirstUser()
	if u == nil {
		slog.Error("failed to get user info from context")
		return nil
	}

	return h.WhoAreUser(b, ctx, u.ID)
}

func (h *Handler) WhoAreUser(b *gotgbot.Bot, ctx *ext.Context, userID int64) error {

	m, err := h.service.GetMemberStats(ctx.EffectiveChat.Id, userID)
	if err != nil {
		slog.Error("failed to get member stats", "chat_id", ctx.EffectiveChat.Id, "error", err)
		return nil
	}

	customTitle := "—"
	if m.CustomTitle != nil && *m.CustomTitle != "" {
		customTitle = *m.CustomTitle
	}
	restText := "—"
	if m.RestUntil != nil {
		restText = " • Рест до " + helpers.FormatToHumanDate(*m.RestUntil)
	}

	text := fmt.Sprintf(
		`<b>📊 Инфомрация о пользователе %s</b>

⚡ <b>Профиль</b>
 • Имя: %s
 • Роль: %s
 • Статус: %s
 • Присоединился: %s

🌟 <b>Активность</b>
 • Сегодня: %d сообщений
 • Календарная неделя: %d сообщений
 • Неделя: %d сообщений
 • Месяц: %d сообщений
 • Всего: %d сообщений

💤 <b>Рест</b>
%s
`,
		helpers.LinkWithContent(m.User, fmt.Sprintf("%s (%s)", m.User.FirstName, customTitle)),
		htmlEscape(m.User.FirstName),
		htmlEscape(customTitle),
		htmlEscape(m.Status),
		helpers.FormatToHumanDate(m.JoinedAt),
		m.DayCount,
		m.WeekCount,
		m.WeekRollingCount,
		m.MonthCount,
		m.AllTime,
		restText,
	)

	_, err = b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})
	if err != nil {
		slog.Error("failed to send message", "chat_id", ctx.EffectiveChat.Id, "error", err)
	}

	return nil
}

func htmlEscape(s string) string {
	return html.EscapeString(s)
}

func formatReport(report []model.MessageReportMember, restMembers []model.RestMember, from, to *time.Time) string {
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
