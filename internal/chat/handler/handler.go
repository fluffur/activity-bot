package handler

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/view"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/session"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

type Handler struct {
	service        *chat.Service
	adminService   *admin.Service
	memberService  *member.Service
	sessionService *session.Service
	dateParser     *helpers.DateParser
}

type chatInfo struct {
	ID    int64
	Title string
}

func New(service *chat.Service, adminService *admin.Service, memberService *member.Service, sessionService *session.Service, dateParser *helpers.DateParser) *Handler {
	return &Handler{service, adminService, memberService, sessionService, dateParser}
}

func (h *Handler) ShowNorm(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteNorm(eb, c.NormWarn, c.NormBan)
		return nil
	})), nil)
	return err
}

func (h *Handler) SetNorm(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	normInt, err := ctx.Number()
	if err != nil {
		return err
	}
	norm := int32(normInt)
	action := ctx.TextOrDefault("варн")
	validActions := map[string]bool{
		"варн": true,
		"бан":  true,
	}
	if !validActions[action] {
		_, err := ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("❌ Неверное действие: '%s'. Допустимые: 'варн', 'бан'", action)), nil)
		return err
	}

	if action == "варн" {
		if err := h.service.SetWarnNorm(ctx.StdContext(), c.ID, int(norm)); err != nil {
			return err
		}
	} else {
		if err := h.service.SetBanNorm(ctx.StdContext(), c.ID, int(norm)); err != nil {
			return err
		}
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatNormSet(int(norm), action)), nil)
	return err
}

func (h *Handler) RemoveNorm(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	action := ctx.TextOrDefault("варн")
	var oldNorm int32
	var setFn func() error

	switch action {
	case "варн":
		oldNorm = c.NormWarn
		setFn = func() error {
			return h.service.SetWarnNorm(ctx.StdContext(), c.ID, 0)
		}
	case "бан":
		oldNorm = c.NormBan
		setFn = func() error {
			return h.service.SetBanNorm(ctx.StdContext(), c.ID, 0)
		}
	default:
		_, err := ctx.Reply(u, ext.ReplyTextString("❌ Неверное действие. Допустимые: 'варн', 'бан'"), nil)
		return err
	}

	if oldNorm == 0 {
		_, err := ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("❌ Норма на %s не была установлена", action)), nil)
		return err
	}

	if err := setFn(); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Норма %d (%s) удалена", oldNorm, action)), nil)
	return err
}

func (h *Handler) ShowNewbieThreshold(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatNewbieThreshold(c.NewbieThresholdDays)), nil)
	return err
}
func (h *Handler) SetNewbieThreshold(ctx *command.Context, u *ext.Update) error {
	days := int32(ctx.NumberOrDefault(3))
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	if days < 0 {
		_, err := ctx.Reply(u, ext.ReplyTextString("Дата не может быть в прошлом!"), nil)
		return err
	}

	if err := h.service.SetNewbieThreshold(ctx.StdContext(), c.ID, days); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatNewbieThresholdSet(days)), nil)
	return err
}

func (h *Handler) SetPrompt(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.SetChatPrompt(ctx.StdContext(), c.ID, ctx.RawArgsHTML); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString("Промпт установлен успешно"), nil)
	return err
}

func (h *Handler) ShowPrompt(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString(view.FormatPrompt(c.AISystemPrompt)), nil)
	return err
}

func (h *Handler) ShowWeekStart(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteWeekStart(eb, int(c.WeekStartDay), c.WeekStartTime)
		return nil
	})), nil)
	return err
}

func (h *Handler) SetWeekStart(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	date, err := ctx.Date()
	if err != nil {
		return err
	}
	weekday := date.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	hour := date.Hour()
	minute := date.Minute()
	if err := h.service.SetWeekStartDay(ctx.StdContext(), c.ID, int(weekday)); err != nil {
		return err
	}
	newTime := fmt.Sprintf("%0.2d:%0.2d", hour, minute)
	if err := h.service.SetWeekStartTime(ctx.StdContext(), c.ID, newTime); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WriteWeekStartSet(eb, int(weekday), newTime)
		return nil
	})), nil)
	return err
}

func (h *Handler) ShowPrefix(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WritePrefix(eb, c.CommandPrefix)
		return nil
	})), nil)
	return err
}

func (h *Handler) SetPrefix(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	prefix := ctx.TextOrDefault("")
	if prefix == "" {
		_, err := ctx.Reply(u, ext.ReplyTextString("❌ Укажите префикс"), nil)
		return err
	}

	if err := h.service.SetCommandPrefix(ctx.StdContext(), c.ID, prefix); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WritePrefixSet(eb, prefix)
		return nil
	})), nil)
	return err
}

func (h *Handler) DisablePrefixOnly(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.SetAllowPrefixless(ctx.StdContext(), c.ID, true); err != nil {
		return err
	}
	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WritePrefixlessToggle(eb, true)
		return nil
	})), nil)
	return err
}

func (h *Handler) EnablePrefixOnly(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	if err := h.service.SetAllowPrefixless(ctx.StdContext(), c.ID, false); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WritePrefixlessToggle(eb, false)
		return nil
	})), nil)
	return err
}

func (h *Handler) ShowPrefixlessStatus(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		view.WritePrefixlessStatus(eb, c.AllowPrefixless)
		return nil
	})), nil)
	return err
}

func (h *Handler) Manage(ctx *command.Context, u *ext.Update) error {
	chatIDs, err := h.service.GetUserManagedChats(ctx.StdContext(), u.EffectiveUser().GetID(), h.adminService.OwnerID())
	if err != nil {
		return err
	}

	if len(chatIDs) == 0 {
		return h.SendDM(ctx, u, "❌ У вас нет доступных чатов для управления", nil)
	}

	chats := make([]chatInfo, 0)
	for _, c := range chatIDs {
		title := c.Title
		if title == "" {
			title = "Чат без названия"
		}

		chats = append(chats, chatInfo{ID: c.ID, Title: title})
	}

	if len(chats) == 0 {
		return h.SendDM(ctx, u, "❌ Не удалось получить информацию о ваших чатах", nil)
	}

	return h.SendDM(ctx, u, "Выберите чат для управления:", h.getManageKeyboard(chats, 1))
}

func (h *Handler) SendDM(ctx *command.Context, u *ext.Update, text string, replyMarkup tg.ReplyMarkupClass) error {
	if u.EffectiveChat().GetID() == u.EffectiveUser().GetID() {
		_, err := ctx.Reply(u, ext.ReplyTextString(text), &ext.ReplyOpts{Markup: replyMarkup})
		return err
	}

	_, err := ctx.SendMessage(
		u.EffectiveUser().GetID(),
		&tg.MessagesSendMessageRequest{
			Message:     text,
			ReplyMarkup: replyMarkup,
		},
	)
	if err != nil {
		_, err = ctx.Reply(u, ext.ReplyTextString("Не удалось отправить сообщение в лс"), nil)
		return err
	}
	_, err = ctx.Reply(u, ext.ReplyTextString("Ответ отправлен в лс"), nil)
	return err
}

func (h *Handler) getManageKeyboard(chats []chatInfo, page int) tg.ReplyMarkupClass {
	const itemsPerPage = 8

	totalPages := (len(chats) + itemsPerPage - 1) / itemsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if endIdx > len(chats) {
		endIdx = len(chats)
	}

	var rows []tg.KeyboardButtonRow
	for _, c := range chats[startIdx:endIdx] {
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: c.Title,
					Data: []byte("manage:" + strconv.FormatInt(c.ID, 10)),
				},
			},
		})
	}

	var navButtons []tg.KeyboardButtonClass
	if page > 1 {
		navButtons = append(navButtons, &tg.KeyboardButtonCallback{
			Text: "< Назад",
			Data: []byte("manage_page:" + strconv.Itoa(page-1)),
		})
	}
	if page < totalPages {
		navButtons = append(navButtons, &tg.KeyboardButtonCallback{
			Text: "Вперед >",
			Data: []byte("manage_page:" + strconv.Itoa(page+1)),
		})
	}

	if len(navButtons) > 0 {
		rows = append(rows, tg.KeyboardButtonRow{Buttons: navButtons})
	}

	return &tg.ReplyInlineMarkup{Rows: rows}
}

func (h *Handler) CallbackManagePage(ctx *command.Context, u *ext.Update) error {
	data, _ := u.CallbackQuery.GetData()
	pageStr := strings.TrimPrefix(string(data), "manage_page:")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return err
	}

	chatIDs, err := h.service.GetUserManagedChats(ctx.StdContext(), u.EffectiveUser().GetID(), h.adminService.OwnerID())
	if err != nil {
		return err
	}

	chats := make([]chatInfo, 0)
	for _, c := range chatIDs {
		title := c.Title
		if title == "" {
			title = "Чат без названия"
		}

		chats = append(chats, chatInfo{ID: c.ID, Title: title})
	}

	if len(chats) == 0 {
		_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			Message: "У вас нет доступных чатов для управления",
		})
		return err
	}

	_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		ID:          u.CallbackQuery.GetMsgID(),
		ReplyMarkup: h.getManageKeyboard(chats, page),
	})
	if err != nil && !strings.Contains(err.Error(), "MESSAGE_NOT_MODIFIED") {
		return err
	}

	_, _ = ctx.AnswerCallback(nil)
	return nil
}

func (h *Handler) OnNewChatTitle(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}
	if u.EffectiveMessage.Action == nil {
		return nil
	}
	sc, ok := u.EffectiveMessage.Action.(*tg.MessageActionChatEditTitle)
	if !ok {
		return nil
	}
	newTitle := sc.Title
	if newTitle == "" {
		return nil
	}

	_, err = h.service.EnsureChatExists(ctx.StdContext(), c.ID, newTitle)
	return err
}

func (h *Handler) CallbackManage(ctx *command.Context, u *ext.Update) error {
	data, _ := u.CallbackQuery.GetData()
	chatIDStr := strings.TrimPrefix(string(data), "manage:")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return err
	}
	m, err := h.memberService.GetChatMember(ctx.StdContext(), chatID, u.EffectiveUser().GetID())
	if err != nil {
		return err
	}
	if !m.StatusGranted(model.StatusModerator) && h.adminService.OwnerID() != m.User.ID {
		_, err := ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
			Message: "У вас больше нет прав администратора в этом чате",
			Alert:   true,
		})
		return err
	}

	if err := h.sessionService.SetActiveChat(ctx.StdContext(), u.EffectiveUser().GetID(), chatID); err != nil {
		return err
	}

	title := "Chat"
	chats, err := ctx.Raw.MessagesGetChats(ctx, []int64{chatID})
	if err == nil && len(chats.GetChats()) > 0 {
		c := chats.GetChats()[0]
		switch ch := c.(type) {
		case *tg.Chat:
			title = ch.Title
		case *tg.Channel:
			title = ch.Title
		case *tg.ChatForbidden:
			title = ch.Title
		case *tg.ChannelForbidden:
			title = ch.Title
		}
	}

	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		Message: "Выбран чат: " + title,
	})

	_, err = ctx.EditMessage(u.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		ID:      u.CallbackQuery.GetMsgID(),
		Message: "Теперь вы управляете чатом: **" + title + "**\nВсе команды настроек теперь будут применяться к этому чату.",
		Entities: []tg.MessageEntityClass{
			&tg.MessageEntityBold{Offset: 27, Length: int(len(title))},
		},
	})
	return err
}

func (h *Handler) UserChats(ctx *command.Context, u *ext.Update) error {
	chats, err := h.service.ListChatsWithoutNorm(
		ctx.StdContext(),
		u.EffectiveUser().GetID(),
	)
	if err != nil {
		return err
	}

	if len(chats) == 0 {
		_, err := ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("%s Все недельные нормы выполнены.", helpers.SuccessEmoji())), nil)
		return err
	}

	var warnChats []model.ChatWithoutNorm
	var banChats []model.ChatWithoutNorm

	for _, c := range chats {
		if c.NormBan > 0 && c.WeekCount < int64(c.NormBan) {
			banChats = append(banChats, c)
			continue
		}

		if c.NormWarn > 0 && c.WeekCount < int64(c.NormWarn) {
			warnChats = append(warnChats, c)
		}
	}

	f := func(eb *entity.Builder) error {
		eb.Plain("📉 ")
		eb.Bold("Невыполненные нормы")
		eb.Plain("\n\n")

		if len(banChats) > 0 {
			eb.Plain("🚫 ")
			eb.Bold("Бан")
			eb.Plain("\n")
			token := eb.Token()

			for i, c := range banChats {
				eb.TextURL(c.Title, chatLink(c.ID))
				eb.Plain("\n")

				writeNormInfoEB(eb, c)

				eb.Plain(fmt.Sprintf("Сообщений: %d\n", c.WeekCount))
				if i < len(banChats)-1 {
					eb.Plain("\n")
				}
			}

			token.Apply(eb, entity.Blockquote(true))
		}

		if len(warnChats) > 0 {
			eb.Plain("⚠️ ")
			eb.Bold("Варн")
			eb.Plain("\n")
			token := eb.Token()

			for i, c := range warnChats {
				eb.TextURL(c.Title, chatLink(c.ID))
				eb.Plain("\n")

				writeNormInfoEB(eb, c)

				eb.Plain(fmt.Sprintf("Сообщений: %d\n", c.WeekCount))
				if i < len(warnChats)-1 {
					eb.Plain("\n")
				}
			}

			token.Apply(eb, entity.Blockquote(true))
			eb.Plain("\n")
		}

		eb.Plain("Всего проблемных чатов: ")
		eb.Bold(strconv.Itoa(len(banChats) + len(warnChats)))
		return nil
	}

	eb := &entity.Builder{}
	if err := f(eb); err != nil {
		return err
	}
	msg, entities := eb.Complete()

	_, err = ctx.SendMessage(
		u.EffectiveUser().GetID(),
		&tg.MessagesSendMessageRequest{
			Message:  msg,
			Entities: entities,
		},
	)
	if err != nil {
		if u.EffectiveChat().GetID() != u.EffectiveUser().GetID() {
			_, err = ctx.Reply(u, ext.ReplyTextString("Не удалось отправить список норм в личные сообщения, убедитесь в том, что бот имеет доступ к личным сообщениям"), nil)
			return err
		}

		return err
	}

	if u.EffectiveChat().GetID() != u.EffectiveUser().GetID() {
		_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
			eb.Plain(fmt.Sprintf("Список невыполненных норм отправлен вам в личные сообщения %s", helpers.SuccessEmoji()))
			return nil
		})), nil)
		return err
	}

	return nil
}

func (h *Handler) EnableTags(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.EnableTags(ctx.StdContext(), c.ID, true); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString("Поддержка тегов в чате включена. Теперь при установке роли админка не выдается"), nil)
	return err
}

func (h *Handler) DisableTags(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if err := h.service.EnableTags(ctx.StdContext(), c.ID, false); err != nil {
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString("Поддержка тегов в чате выключена. Теперь при установке роли выдается админка с минимальными правами"), nil)
	return err
}

func (h *Handler) ShowTags(ctx *command.Context, u *ext.Update) error {
	c, err := ctx.Chat()
	if err != nil {
		return err
	}

	if c.TagsEnabled {
		_, err = ctx.Reply(u, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
			eb.Plain(helpers.SuccessEmoji() + " В чате поддерживаются теги. Это значит, что при установке роли пользователю устанавливается телеграм-тег, а не минимальные права администратора с подписью")
			return nil
		})), nil)
		return err
	}

	_, err = ctx.Reply(u, ext.ReplyTextString("❌ В чате не поддерживаются теги. Это значит, что при установке роли пользователю выдаются минимальные права администратора с подписью, а не телеграм-тег"), nil)
	return err
}

func writeNormInfoEB(eb *entity.Builder, c model.ChatWithoutNorm) {
	var normParts []string

	if c.NormWarn > 0 {
		normParts = append(normParts,
			fmt.Sprintf("<%d – варн", c.NormWarn),
		)
	}

	if c.NormBan > 0 {
		normParts = append(normParts,
			fmt.Sprintf("<%d – бан", c.NormBan),
		)
	}

	if len(normParts) > 0 {
		eb.Plain(fmt.Sprintf("Норма: %d (", c.NormWarn))
		eb.Plain(strings.Join(normParts, ", "))
		eb.Plain(")\n")
	}
}

func chatLink(id int64) string {
	s := strconv.FormatInt(id, 10)

	if strings.HasPrefix(s, "-100") {
		internal := strings.TrimPrefix(s, "-100")
		return fmt.Sprintf("https://t.me/c/%s/1", internal)
	}

	return ""
}
