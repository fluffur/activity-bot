package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/rest"
	"activity-bot/internal/stats"
	"activity-bot/internal/stats/view"
	"activity-bot/internal/user"
	"fmt"
	"log/slog"
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
			_, err = ctx.EffectiveMessage.Reply(b, "❌ <b>Неверный формат даты или диапазона.</b>\n\nИспользуйте: <code>01.02-10.02</code>, <code>10</code> (за последние 10 дней), <code>от вчера до сегодня</code> и т.д.", &gotgbot.SendMessageOpts{ParseMode: "HTML"})
			return err
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
		_, err = ctx.EffectiveMessage.Reply(b, "📭 <b>За выбранный период активности не найдено.</b>", &gotgbot.SendMessageOpts{ParseMode: "HTML"})
		return err
	}

	text := view.FormatReport(report, restMembers, from, to)

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

	text := view.FormatProfile(m)

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

	text := view.FormatInactiveMembers(members)

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
