package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/rest"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	"activity-bot/internal/stats/view"
	"activity-bot/internal/user"
	"context"
	"fmt"
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

func ResolveRange(
	dates []time.Time,
	now time.Time,
	weekStartDay int16,
	weekStartTime string,
	loc *time.Location,
) (time.Time, time.Time, error) {

	if len(dates) == 1 {
		parsed := dates[0]
		dur := parsed.Sub(now)

		from := now.Add(-dur)
		to := now
		return from, to, nil
	}

	if len(dates) == 2 {
		return dates[0], dates[1], nil
	}

	var hour, minutes int
	if _, err := fmt.Sscanf(weekStartTime, "%d:%d", &hour, &minutes); err != nil {
		return time.Time{}, time.Time{}, err
	}

	targetWeekday := time.Weekday(weekStartDay)

	diff := int(now.Weekday() - targetWeekday)
	if diff < 0 {
		diff += 7
	}

	from := time.Date(
		now.Year(),
		now.Month(),
		now.Day()-diff,
		hour,
		minutes,
		0,
		0,
		loc,
	)

	to := from.AddDate(0, 0, 7)

	return from, to, nil
}

func (h *Handler) ShowStats(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	now := time.Now().In(helpers.MoscowLocation)

	from, to, err := ResolveRange(
		ctx.Dates(),
		now,
		c.WeekStartDay,
		c.WeekStartTime,
		helpers.MoscowLocation,
	)

	if err != nil {
		return err
	}
	report, err := h.service.GetChatMembersStats(ctx.StdContext(), c.ID, &from, &to)
	if err != nil {
		return err
	}

	restMembers, err := h.restService.GetRestMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}

	if len(report) == 0 && len(restMembers) == 0 {
		return ctx.ReplyHTML(b, "📭 <b>За выбранный период активности не найдено.</b>")
	}

	text := view.FormatStats(report, restMembers, c.NewbieThresholdDays, &from, &to)

	return ctx.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: getCallKeyboard(c),
	})
}

func (h *Handler) ShowChatActivityGraph(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	now := time.Now().In(helpers.MoscowLocation)

	from, to, err := ResolveRange(
		ctx.Dates(),
		now,
		c.WeekStartDay,
		c.WeekStartTime,
		helpers.MoscowLocation,
	)

	buf, err := h.service.GetChatActivityGraph(ctx.StdContext(), c.ID, &from, &to)
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

	caption := fmt.Sprintf("%s <b>Активность чата</b>", helpers.StatsEmoji())
	if !from.IsZero() && !to.IsZero() {
		caption += fmt.Sprintf(
			"\n%s — %s",
			helpers.FormatToHumanDateTime(from),
			helpers.FormatToHumanDateTime(to),
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

func (h *Handler) WhoAmI(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	return h.WhoAreUser(b, ctx.StdContext(), ctx.Context, c.ID, ctx.EffectiveSender.Id())
}

func (h *Handler) WhoAreYou(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	u, err := ctx.UserOrReply()
	if err != nil {
		return err
	}

	return h.WhoAreUser(b, ctx.StdContext(), ctx.Context, c.ID, u.User.ID)
}

func (h *Handler) WhoAreUser(
	b *gotgbot.Bot,
	ctx context.Context,
	tgCtx *ext.Context,
	dataChatID int64,
	userID int64,
) error {

	m, err := h.service.GetChatMemberStats(ctx, dataChatID, userID)
	if err != nil {
		return err
	}

	text := view.FormatProfile(m, false)

	kb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:              "Вся активность",
					CallbackData:      fmt.Sprintf("profile_activity:%d", userID),
					IconCustomEmojiId: "5425112292683435471",
					Style:             "primary",
				},
			},
		},
	}

	_, err = tgCtx.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: kb,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

func (h *Handler) CallbackProfileGraph(b *gotgbot.Bot, ctx *cmd.Context) error {
	var userID int64
	if _, err := fmt.Sscanf(ctx.CallbackQuery.Data, "profile_graph:%d", &userID); err != nil {
		return err
	}

	_, _ = ctx.CallbackQuery.Answer(b, nil)

	chatID := ctx.TargetChatID()

	buf, err := h.service.GetMessageActivityGraph(ctx.StdContext(), chatID, userID)
	if err != nil || buf == nil {
		_, err = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Недостаточно данных для графика",
		})
		return err
	}

	media := gotgbot.InputMediaPhoto{
		Media:     gotgbot.InputFileByReader("activity.png", buf),
		Caption:   ctx.EffectiveMessage.OriginalHTML(),
		ParseMode: gotgbot.ParseModeHTML,
	}

	_, _, err = ctx.CallbackQuery.Message.EditMedia(
		b,
		media,
		&gotgbot.EditMessageMediaOpts{},
	)
	return err
}

func (h *Handler) CallbackAllActivity(b *gotgbot.Bot, ctx *cmd.Context) error {
	var userID int64
	if _, err := fmt.Sscanf(ctx.CallbackQuery.Data, "profile_activity:%d", &userID); err != nil {
		return err
	}
	chatID := ctx.TargetChatID()

	m, err := h.service.GetChatMemberStats(ctx.StdContext(), chatID, userID)
	if err != nil {
		return err
	}

	text := view.FormatProfile(m, true)

	if ctx.EffectiveMessage.Text == "" {
		_, _, err = ctx.EffectiveMessage.EditCaption(b, &gotgbot.EditMessageCaptionOpts{
			Caption:   text,
			ParseMode: gotgbot.ParseModeHTML,
		})
		return err
	}

	if _, _, err = ctx.EffectiveMessage.EditText(b, text, &gotgbot.EditMessageTextOpts{
		ParseMode: gotgbot.ParseModeHTML,
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			ShowAboveText: true,
		},
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{{
				Text:         "📊 Показать график",
				CallbackData: fmt.Sprintf("profile_graph:%d", userID),
				Style:        "primary",
			}}},
		},
	}); err != nil {
		return err
	}

	return nil
}

func (h *Handler) ListInactive(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	members, err := h.service.GetInactiveMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return ctx.Reply(
			b,
			fmt.Sprintf("%s Нет неактивных участников за последние сутки", helpers.SuccessEmoji()),
			nil,
		)
	}

	text := view.FormatInactiveMembers(members)

	return ctx.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeHTML,
		ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "Созвать неактивных", CallbackData: "call_inactive"}},
			},
		},
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
}

func (h *Handler) ShowRestList(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	restMembers, err := h.restService.GetRestMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatRestList(restMembers))
}

func (h *Handler) ShowFailedNorm(b *gotgbot.Bot, ctx *command.Context) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	now := time.Now().In(helpers.MoscowLocation)
	from, to, err := ResolveRange(ctx.Dates(), now, c.WeekStartDay, c.WeekStartTime, now.Location())
	if err != nil {
		return ctx.ReplyHTML(b, "❌ <b>Неверный формат даты или диапазона.</b>")
	}

	report, err := h.service.GetChatMembersStats(ctx.StdContext(), c.ID, &from, &to)
	if err != nil {
		return err
	}

	text := view.FormatFailedNorm(report, &from, &to)

	if len(report) == 0 {
		return ctx.ReplyHTML(b, text)
	}

	return ctx.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeHTML,
		ReplyMarkup: getCallKeyboard(c),
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})
}

func getCallKeyboard(c model.Chat) *gotgbot.InlineKeyboardMarkup {
	var buttons [][]gotgbot.InlineKeyboardButton
	var row []gotgbot.InlineKeyboardButton

	if c.NormWarn != 0 {
		row = append(row, gotgbot.InlineKeyboardButton{
			Text:              fmt.Sprintf("Без нормы %d", c.NormWarn),
			CallbackData:      "call_no_norm_warn",
			IconCustomEmojiId: "5433866857666855412",
		})
	}
	if c.NormBan != 0 {
		row = append(row, gotgbot.InlineKeyboardButton{
			Text:              fmt.Sprintf("Без нормы %d", c.NormBan),
			CallbackData:      "call_no_norm_ban",
			IconCustomEmojiId: "5433866857666855412",
		})
	}

	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	if c.NormWarn != 0 || c.NormBan != 0 {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: "Всех без нормы", CallbackData: "call_no_norm", IconCustomEmojiId: "5433866857666855412"},
		})
	}

	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}
}

func (h *Handler) ShowNewbies(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.TargetChatID())
	if err != nil {
		return err
	}

	from, to, err := h.resolvePeriod(ctx, time.Weekday(c.WeekStartDay), c.WeekStartTime)
	if err != nil {
		return ctx.ReplyHTML(b, "❌ <b>Неверный формат даты или диапазона.</b>")
	}

	report, err := h.service.GetChatMembersStats(ctx.StdContext(), ctx.TargetChatID(), from, to)
	if err != nil {
		return err
	}

	return ctx.ReplyHTML(b, view.FormatNewbies(report))
}

func (h *Handler) resolvePeriod(ctx *cmd.Context, weekStartDay time.Weekday, weekStartTime string) (*time.Time, *time.Time, error) {
	if len(ctx.ParsedDates()) > 0 {
		dates := ctx.ParsedDates()
		if len(dates) >= 2 {
			from := dates[0]
			to := dates[1]
			if from.After(to) {
				from, to = to, from
			}
			to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())
			return &from, &to, nil
		}
		from := dates[0]
		return &from, nil, nil
	}

	period := "неделя"
	if len(ctx.Args()) > 0 {
		period = ctx.FirstArgument()
	}

	switch period {
	case "неделя", "":
		from, to := stats.ResolvePeriod(stats.PeriodWeek, time.Now(), int16(weekStartDay), weekStartTime)
		return from, to, nil
	case "месяц":
		from, to := stats.ResolvePeriod(stats.PeriodMonth, time.Now(), int16(weekStartDay), weekStartTime)
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
