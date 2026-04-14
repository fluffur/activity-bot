package handler

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/options"
	"activity-bot/internal/rest"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	"activity-bot/internal/stats/view"
	"activity-bot/internal/user"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
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

func (h *Handler) ShowStats(ctx *command.Context, u *ext.Update) error {
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
		return ctx.ReplyOnly(u, options.WithText("За выбранный период активности не найдено"))
	}

	eb := &entity.Builder{}
	view.WriteStats(eb, report, restMembers, c.NewbieThresholdDays, &from, &to)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ShowChatActivityGraph(ctx *command.Context, u *ext.Update) error {
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

	buf, err := h.service.GetChatActivityGraph(ctx.StdContext(), c.ID, &from, &to)
	if err != nil {
		return err
	}

	if buf == nil {
		return ctx.ReplyOnly(u, options.WithText("Недостаточно данных для построения графика"))
	}

	eb := &entity.Builder{}
	helpers.WriteStatsEmoji(eb)
	eb.Plain(" Активность чата\n")
	if !from.IsZero() {
		helpers.FormattedDate(eb, from)
		eb.Plain(" — ")
		helpers.FormattedDate(eb, to)
	}
	caption, entities := eb.Complete()
	f, err := uploader.NewUploader(ctx.Raw).FromBytes(ctx, "graph.png", buf.Bytes())
	if err != nil {
		return err
	}
	_, err = ctx.SendMedia(
		u.EffectiveChat().GetID(),
		&tg.MessagesSendMediaRequest{
			Message:  caption,
			Entities: entities,
			Media: &tg.InputMediaUploadedPhoto{
				File: f,
			},
		},
	)

	return err
}

func (h *Handler) WhoAmI(ctx *command.Context, u *ext.Update) error {
	return h.WhoAreUser(ctx, u, u.EffectiveUser().GetID())
}

func (h *Handler) WhoAreYou(ctx *command.Context, u *ext.Update) error {
	target, err := ctx.UserOrReply()
	if err != nil {
		return err
	}

	return h.WhoAreUser(ctx, u, target.User.ID)
}

func (h *Handler) WhoAreUser(ctx *command.Context, u *ext.Update, userID int64) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	m, err := h.service.GetChatMemberStats(ctx, c.ID, userID)
	if err != nil {
		return err
	}

	kb := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			{Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: "Вся активность",
					Data: []byte(fmt.Sprintf("profile_activity:%d", userID)),
					Style: tg.KeyboardButtonStyle{
						BgSuccess: true,
						Icon:      5425112292683435471,
					},
				},
			},
			},
		},
	}

	eb := &entity.Builder{}
	view.WriteProfile(eb, m, false)

	return ctx.ReplyOnly(u, options.WithBuilder(eb), options.WithMarkup(kb))
}

func (h *Handler) CallbackProfileGraph(ctx *command.Context, u *ext.Update) error {
	var userID int64
	cq := u.CallbackQuery
	if cq == nil {
		return nil
	}
	data, ok := cq.GetData()
	if !ok {
		return nil
	}
	if _, err := fmt.Sscanf(string(data), "profile_graph:%d", &userID); err != nil {
		return fmt.Errorf("failed to scan callback: %w", err)

	}
	_, _ = ctx.AnswerCallback(nil)

	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	buf, err := h.service.GetMessageActivityGraph(ctx.StdContext(), c.ID, userID)
	if err != nil || buf == nil {
		_, err = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			Message: "Недостаточно данных для графика",
		})
		return err
	}
	f, err := uploader.NewUploader(ctx.Raw).FromBytes(ctx, "graph.png", buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to create graph: %w", err)
	}

	msgID := cq.GetMsgID()
	messages, err := ctx.GetMessages(u.EffectiveChat().GetID(), []tg.InputMessageClass{&tg.InputMessageID{ID: msgID}})
	if err != nil || len(messages) == 0 {
		return err
	}
	m, ok := messages[0].(*tg.Message)
	if !ok {
		return fmt.Errorf("unexpected type %T", messages[0])
	}
	var entities []tg.MessageEntityClass

	if e, ok := m.GetEntities(); ok {
		entities = e
	}
	_, err = ctx.EditMessage(
		u.EffectiveChat().GetID(),
		&tg.MessagesEditMessageRequest{
			ID:          msgID,
			Message:     m.GetMessage(),
			Entities:    entities,
			InvertMedia: false,
			Media: &tg.InputMediaUploadedPhoto{
				File: f,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}
	return nil
}

func (h *Handler) CallbackAllActivity(ctx *command.Context, u *ext.Update) error {
	cq := u.CallbackQuery
	if cq == nil {
		return errors.New("no cq")
	}
	data, ok := cq.GetData()
	if !ok {
		return errors.New("no cq data")
	}
	var userID int64
	if _, err := fmt.Sscanf(string(data), "profile_activity:%d", &userID); err != nil {
		return err
	}
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	m, err := h.service.GetChatMemberStats(ctx.StdContext(), c.ID, userID)
	if err != nil {
		return err
	}

	eb := &entity.Builder{}
	view.WriteProfile(eb, m, true)
	text, entities := eb.Complete()

	_, err = ctx.EditMessage(c.ID, &tg.MessagesEditMessageRequest{
		Message:     text,
		ID:          cq.GetMsgID(),
		Entities:    entities,
		InvertMedia: true,
		ReplyMarkup: &tg.ReplyInlineMarkup{
			Rows: []tg.KeyboardButtonRow{
				{Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonCallback{
						Style: tg.KeyboardButtonStyle{
							BgSuccess: true,
						},
						Text: "📊 Показать график",
						Data: []byte(fmt.Sprintf("profile_graph:%d", userID)),
					},
				}},
			},
		},
	})

	return err
}

func (h *Handler) ListInactive(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	log.Println(ctx.Date())

	members, err := h.service.GetInactiveMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return ctx.ReplyOnly(u, options.WithText("Нет неактивных участников за сутки"))
	}

	eb := &entity.Builder{}
	view.WriteInactiveMembers(eb, members)

	return ctx.ReplyOnly(u, options.WithBuilder(eb), options.WithMarkup(&tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			{Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{Text: "Созвать неактивных", Data: []byte("call_inactive")},
			}},
		},
	}))
}

func (h *Handler) ShowRestList(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	restMembers, err := h.restService.GetRestMembers(ctx.StdContext(), c.ID)
	if err != nil {
		return err
	}

	eb := &entity.Builder{}
	view.WriteRestList(eb, restMembers)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ShowFailedNorm(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	now := time.Now().In(helpers.MoscowLocation)
	from, to, err := ResolveRange(ctx.Dates(), now, c.WeekStartDay, c.WeekStartTime, now.Location())
	if err != nil {
		return fmt.Errorf("resolve range error: %w", err)

	}

	report, err := h.service.GetChatMembersStats(ctx.StdContext(), c.ID, &from, &to)
	if err != nil {
		return fmt.Errorf("get chat member stats error: %w", err)
	}

	eb := &entity.Builder{}
	view.WriteFailedNorm(eb, report, &from, &to)

	return ctx.ReplyOnly(u, options.WithBuilder(eb), options.WithMarkup(getCallKeyboard(c)))
}

func getCallKeyboard(c model.Chat) tg.ReplyMarkupClass {
	var buttons []tg.KeyboardButtonClass
	var rows []tg.KeyboardButtonRow

	if c.NormWarn != 0 {
		buttons = append(buttons, &tg.KeyboardButtonCallback{
			Text: fmt.Sprintf("Без нормы %d", c.NormWarn),
			Data: []byte("call_no_norm_warn"),
			Style: tg.KeyboardButtonStyle{
				Icon: 5433866857666855412,
			},
		})
	}
	if c.NormBan != 0 {
		buttons = append(buttons, &tg.KeyboardButtonCallback{
			Text: fmt.Sprintf("Без нормы %d", c.NormBan),
			Data: []byte("call_no_norm_ban"),
			Style: tg.KeyboardButtonStyle{
				Icon: 5433866857666855412,
			},
		})
	}

	if len(buttons) > 0 {
		rows = append(rows, tg.KeyboardButtonRow{Buttons: buttons})
	}

	if c.NormWarn != 0 || c.NormBan != 0 {
		rows = append(rows, tg.KeyboardButtonRow{Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonCallback{
				Text: "Всех без нормы",
				Data: []byte("call_no_norm"),
				Style: tg.KeyboardButtonStyle{
					Icon: 5433866857666855412,
				},
			},
		}})
	}

	return &tg.ReplyInlineMarkup{
		Rows: rows,
	}
}

func (h *Handler) ShowNewbies(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	report, err := h.service.GetChatMembersStats(ctx.StdContext(), c.ID, nil, nil)
	if err != nil {
		return err
	}

	eb := &entity.Builder{}
	view.WriteNewbies(eb, report)

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}
