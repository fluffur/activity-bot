package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/rest"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	"activity-bot/internal/stats/view"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service        *stats.Service
	restService    *rest.Service
	memberService  *member.Service
	userService    *user.Service
	chatService    *chat.Service
	sessionService *session.Service
}

func New(service *stats.Service, restService *rest.Service, memberService *member.Service, userService *user.Service, chatService *chat.Service, sessionService *session.Service) *Handler {
	return &Handler{service, restService, memberService, userService, chatService, sessionService}
}

func (h *Handler) ShowStats(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	from, to, err := h.resolvePeriod(ctx, time.Weekday(c.WeekStartDay))
	if err != nil {
		return ctx.ReplyHTML(b, "❌ <b>Неверный формат даты или диапазона.</b>\n\nИспользуйте: <code>01.02-10.02</code>, <code>10</code> (за последние 10 дней), <code>от вчера до сегодня</code> и т.д.")
	}

	report, err := h.service.GetAllMembersStats(ctx.StdContext(), ctx.TargetChatID(), from, to)
	if err != nil {
		return err
	}

	restMembers, err := h.restService.GetRestMembers(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	if len(report) == 0 && len(restMembers) == 0 {
		return ctx.ReplyHTML(b, "📭 <b>За выбранный период активности не найдено.</b>")
	}

	text := view.FormatReport(report, restMembers, from, to)

	return ctx.ReplyHTML(b, text)
}

func (h *Handler) ShowChatActivityGraph(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	from, to, err := h.resolvePeriod(ctx, time.Weekday(c.WeekStartDay))
	if err != nil {
		from, to = stats.ResolvePeriod(stats.PeriodWeek, time.Now().In(helpers.MoscowLocation), c.WeekStartDay)
	}

	buf, err := h.service.GetChatActivityGraph(ctx.StdContext(), ctx.TargetChatID(), from, to)
	if err != nil {
		return err
	}

	if buf == nil {
		return ctx.Reply(
			b,
			"📉 Недостаточно данных для построения графика",
			nil,
		)
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
			Caption: caption,
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                ctx.EffectiveMessage.MessageId,
				ChatId:                   ctx.EffectiveChat.Id,
				AllowSendingWithoutReply: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		},
	)

	return err
}

func (h *Handler) WhoAmI(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.WhoAreUser(b, ctx.StdContext(), ctx.Context, ctx.TargetChatID(), ctx.EffectiveSender.Id())
}

func (h *Handler) WhoAreYou(b *gotgbot.Bot, ctx *cmd.Context) error {
	u := ctx.FirstUser()
	if u == nil {
		role := ctx.FirstArgument()
		if role == "" {
			return fmt.Errorf("no role no user")
		}

		members, err := h.userService.GetByCustomTitle(ctx.StdContext(), ctx.TargetChatID(), role)
		if err != nil || len(members) == 0 {
			return fmt.Errorf("user with role %s not found", role)
		}

		if len(members) == 1 {
			return h.WhoAreUser(b, ctx.StdContext(), ctx.Context, ctx.TargetChatID(), members[0].User.ID)
		}

		var buttons [][]gotgbot.InlineKeyboardButton
		for _, m := range members {
			btn := gotgbot.InlineKeyboardButton{
				Text:         fmt.Sprintf("%s (%s)", m.User.FirstName, m.CustomTitle),
				CallbackData: fmt.Sprintf("whoareyou:%d", m.User.ID),
			}
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{btn})
		}

		kb := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		}

		return ctx.Reply(b, "Выберите пользователя:", &gotgbot.SendMessageOpts{
			ReplyMarkup: kb,
		})
	}

	return h.WhoAreUser(b, ctx.StdContext(), ctx.Context, ctx.TargetChatID(), u.ID)
}

func (h *Handler) CallbackWhoAreYou(b *gotgbot.Bot, ctx *cmd.Context) error {
	var userID int64
	if _, err := fmt.Sscanf(ctx.CallbackQuery.Data, "whoareyou:%d", &userID); err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(b, nil)

	chatID := ctx.TargetChatID()

	var buttons [][]gotgbot.InlineKeyboardButton

	msg := ctx.EffectiveMessage
	if msg == nil || msg.ReplyMarkup == nil {
		return h.WhoAreUser(
			b,
			ctx.StdContext(),
			ctx.Context,
			chatID,
			userID,
		)
	}

	for _, row := range msg.ReplyMarkup.InlineKeyboard {
		var newRow []gotgbot.InlineKeyboardButton

		for _, button := range row {
			var currentUserID int64
			if _, err := fmt.Sscanf(button.CallbackData, "whoareyou:%d", &currentUserID); err != nil {
				continue
			}

			if currentUserID == userID {
				continue
			}

			newRow = append(newRow, button)
		}

		if len(newRow) > 0 {
			buttons = append(buttons, newRow)
		}
	}

	if len(buttons) == 0 {
		if _, _, err := ctx.CallbackQuery.Message.EditText(b, "✅ Выбраны все участники из списка", nil); err != nil {
			return err
		}
	} else {
		if _, _, err := ctx.CallbackQuery.Message.EditReplyMarkup(b, &gotgbot.EditMessageReplyMarkupOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},
		}); err != nil {
			return err
		}
	}

	return h.WhoAreUser(
		b,
		ctx.StdContext(),
		ctx.Context,
		chatID,
		userID,
	)
}

func (h *Handler) WhoAreUser(b *gotgbot.Bot, ctx context.Context, tgCtx *ext.Context, dataChatID int64, userID int64) error {
	m, err := h.service.GetMemberStats(ctx, dataChatID, userID)
	if err != nil {
		return err
	}

	buf, err := h.service.GetMessageActivityGraph(ctx, dataChatID, userID)
	if err != nil {
		slog.Warn("Failed to get graph", "error", err)
	}

	text := view.FormatProfile(m)

	if buf == nil {
		_, err = tgCtx.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		})
		return err
	}

	_, err = b.SendPhoto(tgCtx.EffectiveChat.Id, gotgbot.InputFileByReader("activity.png", buf), &gotgbot.SendPhotoOpts{
		Caption: text,
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId:                tgCtx.EffectiveMessage.MessageId,
			ChatId:                   tgCtx.EffectiveChat.Id,
			AllowSendingWithoutReply: true,
		},
		ParseMode: gotgbot.ParseModeHTML,
	})
	return err
}

func (h *Handler) Inactive(b *gotgbot.Bot, ctx *cmd.Context) error {
	members, err := h.service.GetInactiveMembers(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return ctx.Reply(
			b,
			"✅ Нет неактивных участников за последние сутки",
			nil,
		)
	}

	text := view.FormatInactiveMembers(members)

	return ctx.ReplyHTML(b, text)
}

func (h *Handler) ShowRestList(b *gotgbot.Bot, ctx *cmd.Context) error {
	restMembers, err := h.restService.GetRestMembers(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatRestList(restMembers))
}

func (h *Handler) ShowFailedNorm(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	from, to, err := h.resolvePeriod(ctx, time.Weekday(c.WeekStartDay))
	if err != nil {
		return ctx.ReplyHTML(b, "❌ <b>Неверный формат даты или диапазона.</b>")
	}

	report, err := h.service.GetAllMembersStats(ctx.StdContext(), ctx.TargetChatID(), from, to)
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatFailedNorm(report, from, to))
}

func (h *Handler) ShowNewbies(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	from, to, err := h.resolvePeriod(ctx, time.Weekday(c.WeekStartDay))
	if err != nil {
		return ctx.ReplyHTML(b, "❌ <b>Неверный формат даты или диапазона.</b>")
	}

	report, err := h.service.GetAllMembersStats(ctx.StdContext(), ctx.TargetChatID(), from, to)
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatNewbies(report, from, to))
}

func (h *Handler) resolvePeriod(ctx *cmd.Context, weekStartDay time.Weekday) (*time.Time, *time.Time, error) {
	period := "неделя"
	if len(ctx.Args()) > 0 {
		period = ctx.FirstArgument()
	}

	switch period {
	case "неделя", "":
		from, to := stats.ResolvePeriod(stats.PeriodWeek, time.Now().In(helpers.MoscowLocation), int16(weekStartDay))
		return from, to, nil
	case "месяц":
		from, to := stats.ResolvePeriod(stats.PeriodMonth, time.Now().In(helpers.MoscowLocation), int16(weekStartDay))
		return from, to, nil
	case "всё", "все", "всего", "вся":
		return nil, nil, nil
	default:
		dp := helpers.NewDateParser()
		f, t, ok := dp.ParseRange(ctx.Args())
		if ok {
			return f, t, nil
		}
		return nil, nil, fmt.Errorf("invalid format")
	}
}
