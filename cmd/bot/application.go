package main

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/config"
	"activity-bot/internal/info/genshin"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/tghtml"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	fsm "github.com/fluffur/botapi-fsm"
	"github.com/gotd/log/logzap"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/gotd/botapi"
	"github.com/gotd/botapi/storage"
)

type ApplicationState string

const (
	AppStateIdle        ApplicationState = ""
	AppStateAwaitRole   ApplicationState = "await_role"
	AppStateConfirmRole ApplicationState = "confirm_role"
	AppStatePending     ApplicationState = "pending"
)

type AppStateData struct {
	Role string `json:"role"`
}

type RejectState string
type RejectStateData struct {
	UserID int64 `json:"user_id"`
}

const (
	RejectStateIdle               RejectState = ""
	RejectStateAwaitRejectMessage RejectState = "await_reject_message"
)

type Application struct {
	UserID        int64  `json:"user_id"`
	Role          string `json:"role"`
	Username      string `json:"username"`
	ApplicationID int    `json:"application_id"`
}

func applicationKey(userID int64) string {
	return fmt.Sprintf("application:%d", userID)
}

func saveApplication(
	ctx context.Context,
	redisClient *redis.Client,
	app Application,
) error {
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}

	return redisClient.Set(
		ctx,
		applicationKey(app.UserID),
		data,
		24*time.Hour,
	).Err()
}

func getApplication(
	ctx context.Context,
	redisClient *redis.Client,
	userID int64,
) (*Application, error) {
	data, err := redisClient.Get(
		ctx,
		applicationKey(userID),
	).Bytes()

	if err != nil {
		return nil, err
	}

	var app Application

	if err := json.Unmarshal(data, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

func deleteApplication(
	ctx context.Context,
	redisClient *redis.Client,
	userID int64,
) error {
	return redisClient.Del(
		ctx,
		applicationKey(userID),
	).Err()
}

func runApplicationBot(
	ctx context.Context,
	log *zap.Logger,
	cfg config.Config,
	redisClient *redis.Client,
	chatMemberService *chatmember.Service,
) error {
	botKey := cfg.ApplicationBotToken[:8]
	botLog := log.With(
		zap.String("bot_key", botKey),
	)

	sessionsDir := filepath.Join(cfg.StoragePath, "sessions")
	if err := os.MkdirAll(sessionsDir, 0750); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}

	storePath := filepath.Join(
		sessionsDir,
		fmt.Sprintf("application_%s.bbolt", botKey),
	)

	store, err := storage.Open(storePath)
	if err != nil {
		return fmt.Errorf("open storage: %w", err)
	}
	defer func() { _ = store.Close() }()

	appFSM := fsm.NewRedisFSM[
		ApplicationState,
		AppStateData,
	](
		redisClient,
		fmt.Sprintf("fsm:%s:app:", botKey),
		24*time.Hour,
		AppStateIdle,
		fsm.WithStrategy[
			ApplicationState,
			AppStateData,
		](fsm.StrategySender),
	)

	rejectFSM := fsm.NewRedisFSM[
		RejectState,
		RejectStateData,
	](
		redisClient,
		fmt.Sprintf("fsm:%s:app:reject:", botKey),
		24*time.Hour,
		RejectStateIdle,
	)

	var bot *botapi.Bot

	bot, err = botapi.New(cfg.ApplicationBotToken, botapi.Options{
		AppID:   cfg.AppID,
		AppHash: cfg.AppHash,
		Logger:  logzap.New(botLog),
		Storage: store,
		OnStart: func(ctx context.Context) {
			botLog.Info("Application bot started")
		},
		FloodWait: true,
	})
	if err != nil {
		return fmt.Errorf("create application bot: %w", err)
	}

	bot.OnCallbackQuery(
		func(c *botapi.Context) error {
			cq := c.Update.CallbackQuery
			if cq == nil {
				return nil
			}

			if err := appFSM.Enter(
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
				log.Error("remove old keyboard", zap.Error(err))
			}

			_, err = c.Bot.SendMessage(
				c,
				botapi.ID(cq.From.ID),
				"Хорошо, отправьте желаемую роль заново\n\n"+
					tghtml.PatPatEmoji()+" "+
					tghtml.Link(cfg.RolesPostLink, "Роли флуда"),
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
		},
		botapi.CallbackData("app:new"),
	)

	bot.OnCommand(
		"start",
		"Запустить бота",
		func(c *botapi.Context) error {
			sess, ok, err := appFSM.Get(c)
			if err != nil {
				return err
			}
			if ok && sess.State == AppStatePending {
				_, err := c.Reply("Пожалуйста, подождите пока вашу заявку обработают, перед тем как отправлять еще одну")
				return err
			}

			if err := appFSM.Enter(
				c,
				AppStateAwaitRole,
				AppStateData{},
			); err != nil {
				return err
			}

			_, err = c.Reply("Отправьте этому боту желаемую роль одним сообщением\n\n"+
				tghtml.PatPatEmoji()+" "+tghtml.Link(cfg.RolesPostLink, "Роли флуда"),
				botapi.WithParseMode(botapi.ParseModeHTML),
				botapi.DisableWebPagePreview(),
			)

			return err
		},
		botapi.ChatTypeIs(botapi.ChatTypePrivate),
	)

	bot.OnMessage(
		func(c *botapi.Context) error {
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

			members, err := chatMemberService.ListHumanPresentChatMembers(
				c.Background(),
				cfg.TargetChatID,
			)

			if err != nil {
				log.Error(
					"Failed to list chat members",
					zap.Error(err),
				)

				_, err := c.Reply(
					"Произошла ошибка при проверке роли.",
				)

				return err
			}

			if !hasRole(role) {
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

			if err := appFSM.Enter(
				c,
				AppStateConfirmRole,
				AppStateData{
					Role: role,
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
		},
		appFSM.State(AppStateAwaitRole),
		botapi.HasText(),
		botapi.ChatTypeIs(botapi.ChatTypePrivate),
	)

	bot.OnCallbackQuery(
		func(c *botapi.Context) error {
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

			sess, ok, err := appFSM.Get(c)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			app := Application{
				UserID:   senderID,
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
				sess.Data.Role,
				userRef,
			)

			sent, err := c.Bot.SendMessage(
				c,
				botapi.ID(cfg.ApplicationChatID),
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
			app.ApplicationID = sent.MessageID

			if err := appFSM.Enter(
				c,
				AppStatePending,
				AppStateData{},
			); err != nil {
				return err
			}

			if err := saveApplication(c, redisClient, app); err != nil {
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
		},
		botapi.CallbackData("app:confirm_rules"),
		appFSM.State(AppStateConfirmRole),
	)

	bot.OnCallbackQuery(
		func(c *botapi.Context) error {
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
				return err
			}

			if err != nil {
				return err
			}

			application, err := getApplication(c, redisClient, userID)
			if err != nil {
				return err
			}

			_, err = c.Bot.SendMessage(
				c,
				botapi.ID(application.UserID),
				fmt.Sprintf(
					"Ваша заявка принята, добро пожаловать!\n%s",
					cfg.TargetChatLink,
				),
			)

			if err != nil {
				log.Error(
					"notify applicant",
					zap.Error(err),
				)
			}

			if err := appFSM.ClearByKey(
				c.Background(),
				userID,
			); err != nil {
				return err
			}

			if err := deleteApplication(c, redisClient, application.UserID); err != nil {
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
		},
		botapi.CallbackPrefix("app:accept:"),
	)

	bot.OnCallbackQuery(
		func(c *botapi.Context) error {
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

			sess, ok, err := appFSM.GetByKey(
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

			if err := rejectFSM.Enter(
				c,
				RejectStateAwaitRejectMessage,
				RejectStateData{UserID: userID},
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
		},
		botapi.CallbackPrefix("app:reject:"),
	)

	bot.OnMessage(
		func(c *botapi.Context) error {
			msg := c.Message()
			if msg == nil {
				return nil
			}

			sess, ok, err := rejectFSM.Get(c)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			application, err := getApplication(c, redisClient, sess.Data.UserID)
			if err != nil {
				return err
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
				log.Error(
					"notify applicant rejection",
					zap.Error(err),
				)
			}

			if err := appFSM.ClearByKey(
				c.Background(),
				application.UserID,
			); err != nil {
				return err
			}

			if err := rejectFSM.Clear(c); err != nil {
				return err
			}

			_, err = c.Reply(
				"Заявка отклонена.",
			)

			return err
		},
		rejectFSM.State(RejectStateAwaitRejectMessage),
		botapi.HasText(),
		botapi.Not(botapi.ChatTypeIs(botapi.ChatTypePrivate)),
	)

	botLog.Info("Starting application bot listener")

	return bot.Run(ctx)
}

func hasRole(role string) bool {
	for _, cat := range genshin.Categories {
		for _, rol := range cat.Roles {
			if strings.EqualFold(predicate.NormalizeTag(rol), role) {
				return true
			}
		}
	}
	return false
}
