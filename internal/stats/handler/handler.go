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
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
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
			return ctx.ReplyHTML(b, "❌ <b>Неверный формат даты или диапазона.</b>\n\nИспользуйте: <code>01.02-10.02</code>, <code>10</code> (за последние 10 дней), <code>от вчера до сегодня</code> и т.д.")
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
		return ctx.ReplyHTML(b, "📭 <b>За выбранный период активности не найдено.</b>")
	}

	text := view.FormatReport(report, restMembers, from, to)

	return ctx.ReplyHTML(b, text)
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
			Caption:   caption,
			ParseMode: gotgbot.ParseModeHTML,
		},
	)

	return err
}

func (h *Handler) WhoAmI(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.WhoAreUser(b, ctx.StdContext(), ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
}

func (h *Handler) WhoAreYou(b *gotgbot.Bot, ctx *cmd.Context) error {
	u := ctx.FirstUser()
	if u == nil {
		role := ctx.FirstArgument()
		if role == "" {
			return fmt.Errorf("no role no user")
		}

		users, err := h.userService.GetByCustomTitle(ctx.StdContext(), ctx.EffectiveChat.Id, role)
		if err != nil || len(users) == 0 {
			return fmt.Errorf("user with role %s not found", role)
		}

		if len(users) == 1 {
			return h.WhoAreUser(b, ctx.StdContext(), ctx.EffectiveChat.Id, users[0].User.ID)
		}

		var buttons [][]gotgbot.InlineKeyboardButton
		for _, u := range users {
			btn := gotgbot.InlineKeyboardButton{
				Text:         fmt.Sprintf("%s (%s)", u.User.FirstName, u.CustomTitle),
				CallbackData: fmt.Sprintf("whoareyou:%d", u.User.ID),
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

	return h.WhoAreUser(b, ctx.StdContext(), ctx.EffectiveChat.Id, u.ID)
}

func (h *Handler) CallbackWhoAreYou(b *gotgbot.Bot, ctx *ext.Context) error {
	var userID int64
	if _, err := fmt.Sscanf(ctx.CallbackQuery.Data, "whoareyou:%d", &userID); err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(b, nil)

	chatID := ctx.EffectiveChat.Id

	var buttons [][]gotgbot.InlineKeyboardButton

	msg := ctx.EffectiveMessage
	if msg == nil || msg.ReplyMarkup == nil {
		return h.WhoAreUser(
			b,
			context.Background(),
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
		context.Background(),
		chatID,
		userID,
	)

}

func (h *Handler) WhoAreUser(b *gotgbot.Bot, ctx context.Context, chatID int64, userID int64) error {
	m, err := h.service.GetMemberStats(ctx, chatID, userID)
	if err != nil {
		return err
	}

	buf, err := h.service.GetMessageActivityGraph(ctx, chatID, userID)
	if err != nil {
		slog.Warn("Failed to get graph", "error", err)
	}

	text := view.FormatProfile(m)

	if buf == nil {
		_, err = b.SendMessage(chatID, text, &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		})
		return err
	}

	_, err = b.SendPhoto(chatID, gotgbot.InputFileByReader("activity.png", buf), &gotgbot.SendPhotoOpts{
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
		return ctx.Reply(
			b,
			"✅ Нет неактивных участников за последние сутки",
			nil,
		)
	}

	text := view.FormatInactiveMembers(members)

	return ctx.ReplyHTML(b, text)
}
