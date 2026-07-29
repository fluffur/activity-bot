package application

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/info/genshin"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strconv"
	"strings"

	fsm "github.com/fluffur/botapi-fsm"

	"github.com/gotd/botapi"
)

type Handler struct {
	appFSM    *fsm.Machine[State, AppStateData]
	rejectFSM *fsm.Machine[RejectState, RejectStateData]

	chatMemberService *chatmember.Service
	repository        *Repository

	targetChatID      int64
	targetChatLink    string
	applicationChatID int64
	rolesPostLink     string
}

func NewHandler(
	appFSM *fsm.Machine[State, AppStateData],
	rejectFSM *fsm.Machine[RejectState, RejectStateData],
	chatMemberService *chatmember.Service,
	repository *Repository,
	targetChatID int64,
	applicationChatID int64,
	targetChatLink string,
	rolesPostLink string,
) *Handler {
	return &Handler{
		appFSM:            appFSM,
		rejectFSM:         rejectFSM,
		chatMemberService: chatMemberService,
		targetChatID:      targetChatID,
		applicationChatID: applicationChatID,
		rolesPostLink:     rolesPostLink,
		targetChatLink:    targetChatLink,
		repository:        repository,
	}
}

func (h *Handler) Start(c *botapi.Context) error {
	sess, ok, err := h.appFSM.Get(c)
	if err != nil {
		return err
	}
	if ok && sess.State == AppStatePending {
		_, err := c.Reply("Пожалуйста, подождите пока вашу заявку обработают, перед тем как отправлять еще одну")
		return err
	}

	if err := h.appFSM.Enter(
		c,
		AppStateAwaitRole,
		AppStateData{},
	); err != nil {
		return err
	}

	_, err = c.Reply("Отправьте этому боту желаемую роль одним сообщением\n\n"+
		tghtml.PatPatEmoji()+" "+tghtml.Link(h.rolesPostLink, "Роли флуда"),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func (h *Handler) StartCallback(c *botapi.Context) error {
	cq := c.Update.CallbackQuery
	if cq == nil {
		return nil
	}

	if err := h.appFSM.Enter(
		c,
		AppStateAwaitRole,
		AppStateData{},
	); err != nil {
		return err
	}

	_, err := c.Bot.EditMessageReplyMarkup(
		c,
		botapi.ID(cq.Message.Chat.ID),
		cq.Message.MessageID,
		nil,
	)
	if err != nil {
		return fmt.Errorf("remove old keyboard: %w", err)
	}

	_, err = c.Bot.SendMessage(
		c,
		botapi.ID(cq.From.ID),
		"Хорошо, отправьте желаемую роль заново\n\n"+
			tghtml.PatPatEmoji()+" "+
			tghtml.Link(h.rolesPostLink, "Роли флуда"),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	if err != nil {
		return err
	}

	return c.AnswerCallback(
		botapi.WithCallbackText(
			"Можно отправить новую заявку",
		),
	)
}

func (h *Handler) ProcessRole(c *botapi.Context) error {
	msg := c.Message()
	if msg == nil {
		return nil
	}

	role := predicate.NormalizeTag(msg.Text)

	if role == "" {
		_, err := c.Reply(
			"Пожалуйста, укажите корректную роль.",
		)
		return err
	}

	members, err := h.chatMemberService.ListHumanPresentChatMembers(
		c.Background(),
		h.targetChatID,
	)

	if err != nil {
		return fmt.Errorf("list chat members: %w", err)
	}
	_, foundRole, ok := genshin.FindRole(role)
	if !ok {
		_, err := c.Reply("Данная роль не найдена, пожалуйста укажите в сообщении сушествующую роль")
		return err
	}

	for _, m := range members {
		if strings.EqualFold(predicate.NormalizeTag(m.Tag), role) {
			_, err := c.Reply(
				"Эта роль уже занята. Пожалуйста, выберите другую.",
			)

			return err
		}
	}

	if err := h.appFSM.Enter(
		c,
		AppStateConfirmRole,
		AppStateData{
			Role: foundRole,
		},
	); err != nil {
		return err
	}

	_, err = c.Reply(
		fmt.Sprintf(
			"Перед отправкой заявки подтвердите, что вы ознакомились с %s",
			tghtml.Link("https://telegra.ph/Pravila-fluda-07-23-35", "правилами флуда"),
		),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
		botapi.WithReplyMarkup(
			botapi.InlineKeyboard(
				botapi.InlineRow(
					botapi.InlineButtonData(
						"✅ Подтвердить",
						"app:confirm_rules",
					),
				),
			),
		),
	)

	return err
}

func (h *Handler) ConfirmRules(c *botapi.Context) error {
	cq := c.Update.CallbackQuery
	if cq == nil {
		return nil
	}

	msg := cq.Message
	if msg == nil {
		return nil
	}
	var senderID int64
	var username, firstname, lastname string
	if sender := c.Sender(); sender != nil {
		username = sender.Username
		firstname = sender.FirstName
		lastname = sender.LastName
		senderID = sender.ID
	} else {
		username = msg.Chat.Username
		firstname = msg.Chat.FirstName
		lastname = msg.Chat.LastName
		senderID = msg.Chat.ID
	}

	var userRef string

	if username != "" {
		userRef = "@" + username
	} else {
		userRef = strings.TrimSpace(
			firstname + " " + lastname,
		)

		if userRef == "" {
			userRef = fmt.Sprintf(
				"ID: %d",
				senderID,
			)
		}
	}

	sess, ok, err := h.appFSM.Get(c)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	app := Application{
		UserID:   senderID,
		ChatID:   h.targetChatID,
		Role:     sess.Data.Role,
		Username: username,
	}

	appID := strconv.FormatInt(
		senderID,
		10,
	)

	adminMsg := fmt.Sprintf(
		"Новая заявка на вступление!\n\n"+
			"Роль: %s\n"+
			"Пользователь: %s",
		sess.Data.Role.Name,
		userRef,
	)

	_, err = c.Bot.SendMessage(
		c,
		botapi.ID(h.applicationChatID),
		adminMsg,
		botapi.WithReplyMarkup(
			botapi.InlineKeyboard(
				botapi.InlineRow(
					botapi.InlineButtonData(
						"Принять",
						"app:accept:"+appID,
					),
					botapi.InlineButtonData(
						"Отклонить",
						"app:reject:"+appID,
					),
				),
			),
		),
	)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	if err := h.appFSM.Enter(
		c,
		AppStatePending,
		AppStateData{},
	); err != nil {
		return err
	}

	if err := h.repository.Save(c, app); err != nil {
		return err
	}

	chatID, _ := c.Chat()
	_, err = c.Bot.SendMessage(
		c,
		chatID,
		"Ваша заявка отправлена на рассмотрение. Ожидайте скорого ответа",
	)

	_, _ = c.Bot.EditMessageReplyMarkup(
		c,
		botapi.ID(msg.Chat.ID),
		msg.MessageID,
		nil,
	)

	return err
}

func (h *Handler) Accept(c *botapi.Context) error {
	cq := c.Update.CallbackQuery
	if cq == nil {
		return nil
	}

	prefix := "app:accept:"
	userID, err := strconv.ParseInt(
		strings.TrimPrefix(cq.Data, prefix),
		10,
		64,
	)

	if err != nil {
		return fmt.Errorf("accept: %w", err)
	}

	application, err := h.repository.Get(c, h.targetChatID, userID)
	if err != nil {
		return err
	}
	if application == nil {
		return c.AnswerCallback()
	}

	_, err = c.Bot.SendMessage(
		c,
		botapi.ID(application.UserID),
		fmt.Sprintf(
			"Ваша заявка принята, добро пожаловать!\n%s",
			h.targetChatLink,
		),
	)

	if err != nil {
		return fmt.Errorf("notify applicant: %w", err)
	}

	if err := h.appFSM.ClearByKey(
		c.Background(),
		userID,
	); err != nil {
		return err
	}

	chatID, _ := c.Chat()

	if _, err := c.Bot.EditMessageReplyMarkup(
		c,
		botapi.ID(cq.Message.Chat.ID),
		cq.Message.MessageID,
		nil,
	); err != nil {
		return err
	}

	if _, err := c.Bot.SendMessage(c, chatID, "Заявка успешно принята", botapi.ReplyTo(cq.Message.MessageID)); err != nil {
		return err
	}

	return c.AnswerCallback(
		botapi.WithCallbackText(
			"Заявка принята",
		),
	)
}

func (h *Handler) Reject(c *botapi.Context) error {
	cq := c.Update.CallbackQuery
	if cq == nil {
		return nil
	}

	userID, err := strconv.ParseInt(
		strings.TrimPrefix(cq.Data, "app:reject:"),
		10,
		64,
	)

	if err != nil {
		return err
	}

	sess, ok, err := h.appFSM.GetByKey(
		c.Background(),
		userID,
	)

	if err != nil {
		return err
	}

	if !ok || sess.State != AppStatePending {
		_ = c.AnswerCallback(
			botapi.WithCallbackText(
				"Заявка не найдена или устарела",
			),
		)

		return nil
	}

	if err := h.rejectFSM.Enter(
		c,
		RejectStateAwaitRejectMessage,
		RejectStateData{UserID: userID, ChatID: h.targetChatID},
	); err != nil {
		return err
	}

	chatID, _ := c.Chat()
	_, _ = c.Bot.EditMessageReplyMarkup(
		c,
		botapi.ID(cq.Message.Chat.ID),
		cq.Message.MessageID,
		nil,
	)
	_, err = c.Bot.SendMessage(
		c,
		chatID,
		"Введите причину отказа:",
	)

	return err
}

func (h *Handler) RejectMessage(c *botapi.Context) error {
	msg := c.Message()
	if msg == nil {
		return nil
	}

	sess, ok, err := h.rejectFSM.Get(c)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	application, err := h.repository.Get(c, sess.Data.ChatID, sess.Data.UserID)
	if err != nil {
		return err
	}
	if application == nil {
		return nil
	}
	reason := strings.TrimSpace(msg.Text)

	if reason == "" {
		_, err := c.Reply(
			"Причина не может быть пустой.",
		)

		return err
	}

	_, err = c.Bot.SendMessage(
		c,
		botapi.ID(application.UserID),
		fmt.Sprintf(
			"К сожалению, ваша заявка была отклонена.\n\nПричина: %s",
			reason,
		),
		botapi.WithReplyMarkup(
			botapi.InlineKeyboard(
				botapi.InlineRow(
					botapi.InlineButtonData(
						"📨 Отправить новую заявку",
						"app:new",
					),
				),
			),
		),
	)

	if err != nil {
		return fmt.Errorf("notify applicant rejection: %w", err)
	}

	if err := h.appFSM.ClearByKey(
		c.Background(),
		application.UserID,
	); err != nil {
		return err
	}

	if err := h.rejectFSM.Clear(c); err != nil {
		return err
	}

	_, err = c.Reply(
		"Заявка отклонена.",
	)

	if err := h.repository.Delete(
		c,
		application.ChatID,
		application.UserID,
	); err != nil {
		return err
	}

	return err
}
