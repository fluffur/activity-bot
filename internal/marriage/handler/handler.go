package handler

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/marriage"
	"activity-bot/internal/member"
	"activity-bot/internal/options"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type Handler struct {
	service       *marriage.Service
	memberService *member.Service
}

func New(service *marriage.Service, memberService *member.Service) *Handler {
	return &Handler{
		service:       service,
		memberService: memberService,
	}
}

func (h *Handler) RequestMarriage(ctx *command.Context, u *ext.Update) error {
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}

	user, err := ctx.UserOrReply()
	if err != nil {
		return err
	}

	outcome, err := h.service.HandleMarriageRequest(
		ctx.StdContext(),
		sender.ChatID,
		sender.User.ID,
		user.User.ID,
		false,
	)
	if err != nil {
		switch {
		case errors.Is(err, marriage.ErrAlreadyMarried):
			return ctx.ReplyOnly(u, options.WithText("Кто-то из вас уже состоит в браке"))
		case errors.Is(err, marriage.ErrRequestExists):
			return ctx.ReplyOnly(u, options.WithText("Между вами уже есть активный запрос брака"))
		case errors.Is(err, marriage.ErrInvalidTarget):
			return ctx.ReplyOnly(u, options.WithText("Некорректная цель для запроса брака"))
		default:
			log.Println(err)
			return ctx.ReplyOnly(u, options.WithText("Не удалось выполнить действие"))
		}
	}

	switch outcome.Type {

	case marriage.OutcomeSelf:
		eb := &entity.Builder{}
		eb.Plain("💍 ")
		helpers.WriteRoleEmojiMention(eb, *sender)
		eb.Plain(fmt.Sprintf(
			" торжественно вступил%s в брак с %s собой.\n\nСамодостаточность наше все 💪",
			helpers.Gendered(sender.User.Gender, "", "а"),
			helpers.Gendered(sender.User.Gender, "самим", "самой"),
		))
		return ctx.ReplyOnly(u, options.WithBuilder(eb))

	case marriage.OutcomeDirect:
		eb := &entity.Builder{}
		eb.Plain("💍 ")
		helpers.WriteRoleEmojiMention(eb, *sender)
		eb.Plain(" торжественно заключил(а) брак с ")
		helpers.WriteRoleEmojiMention(eb, *user)
		eb.Plain(" без ожидания подтверждения.\n\nПусть союз будет долгим и счастливым!")
		return ctx.ReplyOnly(u, options.WithBuilder(eb))

	case marriage.OutcomeAutoAccepted:
		eb := &entity.Builder{}
		eb.Plain("💍 ")
		helpers.WriteRoleEmojiMention(eb, *sender)
		eb.Plain(" и ")
		helpers.WriteRoleEmojiMention(eb, *user)
		eb.Plain(" одновременно сделали предложение и сразу вступили в брак!\n\nВот это синхрон 😏")
		return ctx.ReplyOnly(u, options.WithBuilder(eb))

	case marriage.OutcomeRequestCreated:
		eb := &entity.Builder{}
		helpers.WriteRoleEmojiMention(eb, *sender)
		eb.Plain(fmt.Sprintf(
			" %s предложение руки и сердца ",
			helpers.Gendered(sender.User.Gender, "сделал", "сделала"),
		))
		helpers.WriteRoleEmojiMention(eb, *user)
		eb.Plain("\nПринять или отклонить?")

		markup := &tg.ReplyInlineMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: []tg.KeyboardButtonClass{
						&tg.KeyboardButtonCallback{
							Text: "Принять",
							Data: []byte(fmt.Sprintf("marriage_accept:%d", sender.User.ID)),
						},
						&tg.KeyboardButtonCallback{
							Text: "Отклонить",
							Data: []byte(fmt.Sprintf("marriage_reject:%d", sender.User.ID)),
						},
					},
				},
			},
		}

		return ctx.ReplyOnly(u, options.WithBuilder(eb), options.WithMarkup(markup))
	}

	return nil
}

func (h *Handler) ShowMarriage(ctx *command.Context, u *ext.Update) error {
	user, err := ctx.AnyUser()
	if err != nil {
		return err
	}

	m, err := h.service.GetMarriage(ctx.StdContext(), user.ChatID, user.User.ID)
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось получить информацию о браке"))
	}
	if m == nil {
		return ctx.ReplyOnly(u, options.WithText("У пользователя нет активного брака"))
	}

	if m.User1.User.ID == m.User2.User.ID {
		eb := &entity.Builder{}
		helpers.WriteRoleEmojiMention(eb, *user)
		eb.Plain(fmt.Sprintf(" в браке с %s собой", helpers.Gendered(user.User.Gender, "самим", "самой")))
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}
	partner := m.User1
	if partner.User.ID == user.User.ID {
		partner = m.User2
	}

	eb := &entity.Builder{}
	helpers.WriteRoleEmojiMention(eb, *user)
	eb.Plain(" в браке с ")
	helpers.WriteRoleEmojiMention(eb, partner)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) AcceptMarriageRequest(ctx *command.Context, u *ext.Update) error {
	if u.CallbackQuery == nil {
		return nil
	}

	data, _ := u.CallbackQuery.GetData()
	fromUserID, err := parseMarriageCallbackUserID(data)
	if err != nil {
		return err
	}
	toUserID := u.EffectiveUser().GetID()
	chat, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.AcceptMarriageRequest(ctx.StdContext(), chat.ID, fromUserID, toUserID); err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Alert:   true,
			Message: "Не удалось принять запрос брака",
		})
		return nil
	}

	fromMember, _ := h.memberService.GetChatMember(ctx.StdContext(), chat.ID, fromUserID)
	toMember, _ := h.memberService.GetChatMember(ctx.StdContext(), chat.ID, toUserID)

	eb := &entity.Builder{}
	eb.Plain("💍 Согласие получено. Брак заключён!")
	text, entities := eb.Complete()

	_, _ = ctx.EditMessage(chat.ID, &tg.MessagesEditMessageRequest{
		ID:          u.CallbackQuery.GetMsgID(),
		Message:     text,
		Entities:    entities,
		ReplyMarkup: &tg.ReplyInlineMarkup{},
	})
	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: u.CallbackQuery.QueryID,
		Message: "Запрос брака принят",
	})

	announce := &entity.Builder{}
	announce.Plain("🎉 ")
	helpers.WriteRoleEmojiMention(announce, fromMember)
	announce.Plain(" и ")
	helpers.WriteRoleEmojiMention(announce, toMember)
	announce.Plain(" официально вступили в брак!\n\nГорько! 💐")
	_ = ctx.ReplyOnly(u, options.WithBuilder(announce))
	return nil
}

func (h *Handler) RejectMarriageRequest(ctx *command.Context, u *ext.Update) error {
	if u.CallbackQuery == nil {
		return nil
	}
	data, _ := u.CallbackQuery.GetData()
	fromUserID, err := parseMarriageCallbackUserID(data)
	if err != nil {
		return err
	}
	actorID := u.EffectiveUser().GetID()
	chat, err := ctx.Chat()
	if err != nil {
		return err
	}

	if actorID == fromUserID {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Alert:   true,
			Message: "Отклонить может только тот, кому сделали предложение",
		})
		return nil
	}

	requestReceiver, err := h.memberService.GetChatMember(ctx.StdContext(), chat.ID, actorID)
	if err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Alert:   true,
			Message: "Отклонить может только адресат запроса",
		})
		return nil
	}
	requestSender, err := h.memberService.GetChatMember(ctx.StdContext(), chat.ID, fromUserID)
	if err != nil || requestReceiver.User.ID == requestSender.User.ID {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Alert:   true,
			Message: "Отклонить может только адресат запроса",
		})
		return nil
	}

	if err := h.service.RejectMarriageRequest(ctx.StdContext(), chat.ID, fromUserID, actorID, false); err != nil {
		_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: u.CallbackQuery.QueryID,
			Alert:   true,
			Message: "Не удалось отклонить запрос брака",
		})
		return nil
	}

	_, _ = ctx.EditMessage(chat.ID, &tg.MessagesEditMessageRequest{
		ID:          u.CallbackQuery.GetMsgID(),
		Message:     "Запрос брака отклонён",
		ReplyMarkup: &tg.ReplyInlineMarkup{},
	})
	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: u.CallbackQuery.QueryID,
		Message: "Запрос брака отклонён",
	})
	return nil
}

func (h *Handler) Divorce(ctx *command.Context, u *ext.Update) error {
	sender, err := ctx.Sender()
	if err != nil {
		return err
	}

	var partnerID int64
	target, err := ctx.User()
	if err == nil {
		partnerID = target.User.ID
	}

	divorced, err := h.service.Divorce(ctx.StdContext(), sender.ChatID, sender.User.ID, partnerID)
	if err != nil {
		switch err {
		case marriage.ErrNoMarriage:
			return ctx.ReplyOnly(u, options.WithText("У вас нет активного брака"))
		case marriage.ErrNotYourMarriage:
			return ctx.ReplyOnly(u, options.WithText("Можно развестись только со своим текущим супругом"))
		default:
			return ctx.ReplyOnly(u, options.WithText("Не удалось оформить развод"))
		}
	}

	partner := divorced.User1
	if partner.User.ID == sender.User.ID {
		partner = divorced.User2
	}

	eb := &entity.Builder{}
	helpers.WriteRoleEmojiMention(eb, *sender)
	if partner.User.ID == sender.User.ID {
		eb.Plain(fmt.Sprintf(" %s с %s собой",
			helpers.Gendered(sender.User.Gender, "развелся", "развелась"),
			helpers.Gendered(sender.User.Gender, "самим", "самой")))
		return ctx.ReplyOnly(u, options.WithBuilder(eb))
	}
	eb.Plain(" больше не в браке с ")
	helpers.WriteRoleEmojiMention(eb, partner)
	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func (h *Handler) ListMarriages(ctx *command.Context, u *ext.Update) error {
	chat, err := ctx.Chat()
	if err != nil {
		return err
	}

	marriages, err := h.service.ListMarriages(ctx.StdContext(), chat.ID)
	if err != nil {
		return ctx.ReplyOnly(u, options.WithText("Не удалось получить список браков"))
	}
	if len(marriages) == 0 {
		return ctx.ReplyOnly(u, options.WithText("В чате пока нет активных браков"))
	}

	grouped := make(map[string][]marriage.Marriage)
	for _, m := range marriages {
		category := marriageCategory(m.MarriedAt)
		grouped[category] = append(grouped[category], m)
	}

	order := []string{
		"Молодожёны",
		"Зелёная свадьба",
		"Ситцевая свадьба",
		"Бумажная свадьба",
		"Кожаная свадьба",
		"Льняная свадьба",
		"Деревянная свадьба",
		"Чугунная свадьба",
		"Медная свадьба",
		"Жестяная свадьба",
		"Фаянсовая свадьба",
		"Розовая свадьба",
	}

	eb := &entity.Builder{}
	eb.Plain("Активные браки:\n\n")

	writeGroup := func(title string, items []marriage.Marriage) {
		if len(items) == 0 {
			return
		}
		eb.Plain(title)
		eb.Plain("\n")
		for i, m := range items {
			eb.Plain(fmt.Sprintf("%d. ", i+1))
			helpers.WriteRoleEmojiMention(eb, m.User1)
			eb.Plain(" ❤ ")
			if m.User1.User.ID == m.User2.User.ID {
				eb.Plain("себя")
			} else {
				helpers.WriteRoleEmojiMention(eb, m.User2)
			}
			if !m.MarriedAt.IsZero() {
				eb.Plain(" — вместе ")
				eb.Plain(helpers.FormatLastSeenPlain(m.MarriedAt))
			}
			eb.Plain("\n")
		}
		eb.Plain("\n")
	}

	for _, category := range order {
		writeGroup(category, grouped[category])
		delete(grouped, category)
	}
	for category, items := range grouped {
		writeGroup(category, items)
	}

	return ctx.ReplyOnly(u, options.WithBuilder(eb))
}

func marriageCategory(marriedAt time.Time) string {
	if marriedAt.IsZero() {
		return "Без даты"
	}

	d := time.Since(marriedAt)
	if d < 30*24*time.Hour {
		return "Молодожёны"
	}

	years := int(d.Hours() / 24 / 365)
	switch years {
	case 0:
		return "Зелёная свадьба"
	case 1:
		return "Ситцевая свадьба"
	case 2:
		return "Бумажная свадьба"
	case 3:
		return "Кожаная свадьба"
	case 4:
		return "Льняная свадьба"
	case 5:
		return "Деревянная свадьба"
	case 6:
		return "Чугунная свадьба"
	case 7:
		return "Медная свадьба"
	case 8:
		return "Жестяная свадьба"
	case 9:
		return "Фаянсовая свадьба"
	case 10:
		return "Розовая свадьба"
	default:
		return fmt.Sprintf("%d+ лет вместе", years)
	}
}

func (h *Handler) isBotUser(ctx *command.Context, userID int64) (bool, error) {
	inputPeer, err := ctx.ResolveInputPeerById(userID)
	if err != nil {
		return false, err
	}
	peerUser, ok := inputPeer.(*tg.InputPeerUser)
	if !ok {
		return false, nil
	}

	users, err := ctx.Raw.UsersGetUsers(ctx, []tg.InputUserClass{
		&tg.InputUser{
			UserID:     peerUser.UserID,
			AccessHash: peerUser.AccessHash,
		},
	})
	if err != nil {
		return false, err
	}
	if len(users) == 0 {
		return false, nil
	}
	tgUser, ok := users[0].(*tg.User)
	if !ok {
		return false, nil
	}
	return tgUser.Bot, nil
}

func parseMarriageCallbackUserID(data []byte) (int64, error) {
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid callback data")
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
