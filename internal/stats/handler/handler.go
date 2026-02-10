package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/rest"
	"activity-bot/internal/stats"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service       *stats.Service
	restService   *rest.Service
	memberService *member.Service
	userService   *user.Service
	chatService   *chat.Service
}

func New(service *stats.Service, restService *rest.Service, memberService *member.Service, userService *user.Service, chatService *chat.Service) *Handler {
	return &Handler{service, restService, memberService, userService, chatService}
}

func (h *Handler) ShowStats(b *gotgbot.Bot, ctx *cmd.Context) error {
	if _, err := h.memberService.SyncChatMembers(ctx.StdContext(), ctx.EffectiveChat.Id); err != nil {
		slog.Warn("failed to auto-update chat members in stats", "chat_id", ctx.EffectiveChat.Id, "error", err)
	}

	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	period := "неделя"
	if len(ctx.Args()) > 0 {
		period = ctx.FirstArgument()
	}

	var from, to *time.Time
	switch period {
	case "неделя", "":
		from, to = stats.ResolvePeriod(stats.PeriodWeek, time.Now(), c.WeekStartDay)
	case "месяц":
		from, to = stats.ResolvePeriod(stats.PeriodMonth, time.Now(), c.WeekStartDay)
	case "всё", "все", "всего", "вся":
		from, to = nil, nil
	default:
		dp := helpers.NewDateParser()
		f, t, ok := dp.ParseRange(ctx.Args())
		slog.Info("stats range parse", "args", ctx.Args(), "from", f, "to", t, "ok", ok)
		if ok {
			from, to = f, t
		} else {
			return errors.New("invalid date range")
		}
	}

	report, err := h.service.GetAllMembersStats(ctx.StdContext(), ctx.EffectiveChat.Id, from, to)
	if err != nil {
		return err
	}

	restMembers, err := h.restService.GetRestMembers(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	if len(report) == 0 && len(restMembers) == 0 {
		return nil
	}

	text := formatReport(report, restMembers, from, to)

	_, err = ctx.EffectiveMessage.Reply(
		b,
		text,
		&gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeHTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
		},
	)
	return err
}

func (h *Handler) ShowChatActivityGraph(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	period := "неделя"
	if len(ctx.Args()) > 0 {
		period = ctx.FirstArgument()
	}

	var from, to *time.Time
	switch period {
	case "неделя":
		from, to = stats.ResolvePeriod(stats.PeriodWeek, time.Now(), c.WeekStartDay)
	case "месяц":
		from, to = stats.ResolvePeriod(stats.PeriodMonth, time.Now(), c.WeekStartDay)
	case "всё", "все", "всего":
		from, to = nil, nil
	default:
		dp := helpers.NewDateParser()
		if f, t, ok := dp.ParseRange(ctx.Args()); ok {
			from, to = f, t
		} else {
			from, to = stats.ResolvePeriod(stats.PeriodWeek, time.Now(), c.WeekStartDay)
		}
	}

	buf, err := h.service.GetChatActivityGraph(ctx.StdContext(), ctx.EffectiveChat.Id, from, to)
	if err != nil {
		return err
	}

	if buf == nil {
		_, err = ctx.EffectiveMessage.Reply(
			b,
			"📉 Недостаточно данных для построения графика",
			nil,
		)
		return err
	}

	caption := "📊 <b>Активность чата</b>"
	if from != nil && to != nil {
		caption += fmt.Sprintf(
			"\n%s — %s",
			helpers.FormatToHumanDate(*from),
			helpers.FormatToHumanDate(*to),
		)
	}

	_, err = b.SendPhoto(
		ctx.EffectiveChat.Id,
		gotgbot.InputFileByReader("chat_activity.png", buf),
		&gotgbot.SendPhotoOpts{
			Caption:   caption,
			ParseMode: gotgbot.ParseModeHTML,
		},
	)

	return err
}

func (h *Handler) WhoAmI(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.WhoAreUser(b, ctx, ctx.EffectiveSender.Id())
}

func (h *Handler) WhoAreYou(b *gotgbot.Bot, ctx *cmd.Context) error {
	u := ctx.FirstUser()
	if u == nil {
		role := ctx.FirstArgument()
		us, err := h.userService.GetByCustomTitle(ctx.StdContext(), ctx.EffectiveChat.Id, role)
		if err != nil {
			return cmd.ErrNoUser
		}
		u = &us
	}

	return h.WhoAreUser(b, ctx, u.ID)
}

func (h *Handler) WhoAreUser(b *gotgbot.Bot, ctx *cmd.Context, userID int64) error {
	m, err := h.service.GetMemberStats(ctx.StdContext(), ctx.EffectiveChat.Id, userID)
	if err != nil {
		return err
	}

	buf, err := h.service.GetMessageActivityGraph(ctx.StdContext(), ctx.EffectiveChat.Id, userID)
	if err != nil {
		slog.Warn("Failed to get graph", "error", err)
	}

	customTitle := "—"
	if m.CustomTitle != nil && *m.CustomTitle != "" {
		customTitle = *m.CustomTitle
	}

	extraText := ""
	if m.RestUntil != nil {
		extraText = fmt.Sprintf("💤 Рест до %s", helpers.FormatToHumanDate(*m.RestUntil))
	} else {
		if m.WeeklyNorm <= m.WeekCount {
			extraText = "<b>✅ Норма</b>"
		} else {
			extraText = "<b>❌ Норма</b>"
		}
		extraText += fmt.Sprintf(": %d/%d за эту неделю", m.WeekCount, m.WeeklyNorm)
	}

	text := fmt.Sprintf(
		`<b>📊 Информация о %s</b>

🌟 Статус: <b>%s</b> | Присоединился: <b>%s</b>

<b>📅 Активность</b>: сегодня <b>%d</b> | неделя <b>%d</b> | месяц <b>%d</b>

<b>🔄 Активность в последние</b>: 24ч <b>%d</b> | 7д <b>%d</b> | 30д <b>%d</b>

📝 <b>Всего сообщений:</b> %d

%s
`,
		helpers.LinkWithContent(m.User, fmt.Sprintf("%s (%s)", m.User.FirstName, customTitle)),
		helpers.TranslateMemberStatus(m.Status),
		helpers.FormatToHumanDate(m.JoinedAt),
		m.DayCount,
		m.WeekCount,
		m.MonthCount,
		m.DayRollingCount,
		m.WeekRollingCount,
		m.MonthRollingCount,
		m.AllTime,
		extraText,
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
		prefix := "•"

		switch i {
		case 0:
			prefix = "🔥"
		case 1, 2:
			prefix = "⚡"
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", prefix, item))
	}
}

func (h *Handler) Inactive(b *gotgbot.Bot, ctx *cmd.Context) error {
	members, err := h.service.GetInactiveMembers(ctx.StdContext(), ctx.EffectiveChat.Id)
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
