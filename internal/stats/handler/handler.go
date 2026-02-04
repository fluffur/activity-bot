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
		return err

	}

	restMembers, err := h.restService.GetRestMembers(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	if len(report) == 0 && len(restMembers) == 0 {
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
		return cmd.ErrNoUser
	}

	return h.WhoAreUser(b, ctx, u.ID)
}

func (h *Handler) WhoAreUser(b *gotgbot.Bot, ctx *ext.Context, userID int64) error {
	m, err := h.service.GetMemberStats(ctx.EffectiveChat.Id, userID)
	if err != nil {
		return err
	}

	buf, err := h.service.GetMessageActivityGraph(ctx.EffectiveChat.Id, userID)
	if err != nil {
		slog.Warn("Failed to get graph", "error", err)
	}

	customTitle := "—"
	if m.CustomTitle != nil && *m.CustomTitle != "" {
		customTitle = *m.CustomTitle
	}

	restText := "—"
	if m.RestUntil != nil {
		restText = "💤 Рест до " + helpers.FormatToHumanDate(*m.RestUntil)
	}

	text := fmt.Sprintf(
		`<b>📊 Информация о пользователе %s</b>

🌟 <b>Профиль</b>
┌───────────────
│ Роль: %s
│ Статус: %s
│ Присоединился: %s
└───────────────

📅 <b>Активность (календарная)</b>
┌───────────────
│ Сегодня: %d сообщений
│ На этой неделе: %s сообщений
│ В этом месяце: %d сообщений
└───────────────

🔄 <b>Активность (rolling)</b>
┌───────────────
│ Последние 24ч: %d сообщений
│ Последние 7 дней: %d сообщений
│ Последние 30 дней: %d сообщений
└───────────────

📝 <b>Всего сообщений</b>
┌───────────────
│ Всего: %d сообщений
└───────────────

%s
`,
		helpers.LinkWithContent(m.User, fmt.Sprintf("%s (%s)", m.User.FirstName, customTitle)),
		htmlEscape(customTitle),
		htmlEscape(m.Status),
		helpers.FormatToHumanDate(m.JoinedAt),
		m.DayCount,
		fmt.Sprintf("%d из нормы в %d", m.WeekCount, m.WeeklyNorm),
		m.MonthCount,
		m.DayRollingCount,
		m.WeekRollingCount,
		m.MonthRollingCount,
		m.AllTime,
		restText,
	)

	if buf == nil {
		_, err = b.SendMessage(ctx.EffectiveChat.Id, text, &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
		return err
	}

	_, err = b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByReader("activity.png", buf), &gotgbot.SendPhotoOpts{
		Caption:   text,
		ParseMode: "HTML",
	})
	return err
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

func (h *Handler) Inactive(b *gotgbot.Bot, ctx *ext.Context, _ *cmd.Context) error {
	members, err := h.service.GetInactiveMembers(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		_, err = ctx.EffectiveMessage.Reply(
			b,
			"✅ Нет неактивных участников за последние сутки",
			nil,
		)
		return err
	}

	var sb strings.Builder
	sb.WriteString("<b>😴 Неактивные участники (более 1 суток)</b>\n\n")

	for i, m := range members {
		sb.WriteString(fmt.Sprintf(
			"%d. %s (%s)",
			i+1,
			helpers.LinkWithContent(
				m.Member.User,
				fmt.Sprintf("%s", m.Member.User.FirstName),
			),
			m.Member.CustomTitle,
		))

		if m.LastActivity != nil {
			sb.WriteString(fmt.Sprintf(
				" — %s (%s)",
				helpers.FormatToHumanDate(*m.LastActivity),
				helpers.FormatLastSeen(*m.LastActivity),
			))
		} else {
			sb.WriteString(" — не писал ни разу")
		}

		sb.WriteString("\n")
	}

	_, err = ctx.EffectiveMessage.Reply(
		b,
		sb.String(),
		&gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		},
	)

	return err
}
