package main

import (
	"activity-bot/internal/config"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/helpers/participant"
	"activity-bot/internal/helpers/tghtml"
	"activity-bot/internal/middleware"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	fsm "github.com/fluffur/botapi-fsm"
	"github.com/gotd/botapi"
	"github.com/gotd/botapi/storage"
	"github.com/gotd/log/logzap"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewDevelopment()
	defer func() { _ = log.Sync() }()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}
	store, err := storage.Open(cfg.StoragePath)
	if err != nil {
		log.Fatal("Failed to open storage", zap.Error(err))
	}
	defer func() { _ = store.Close() }()

	bot, err := botapi.New(cfg.BotToken, botapi.Options{
		AppID:     cfg.AppID,
		AppHash:   cfg.AppHash,
		Logger:    logzap.New(log),
		Storage:   store,
		FloodWait: true,
	})
	if err != nil {
		log.Fatal("Create bot", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	queries := db.New(pool)

	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})

	fsmStore := fsm.NewRedisJSONStore[profileState, profileData](client, "fsm:", 24*time.Hour)

	bot.Use(botapi.Recover(), botapi.Timeout(time.Minute), botapi.Logging(), middleware.Chat(queries))

	registerCommands(bot, fsmStore)

	log.Info("Starting bot")
	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}

type profileState string

const (
	profileIdle      profileState = ""
	profileAwaitName profileState = "name"
	profileAwaitAge  profileState = "age"
)

type profileData struct {
	Name string
}

func registerProfileFSM(bot *botapi.Bot, store fsm.Store[profileState, profileData]) {
	m := fsm.New(store, profileIdle,
		fsm.WithOnCancel(func(c *botapi.Context, _ *fsm.Session[profileState, profileData]) error {
			_, err := c.Reply("Диалог отменён.")
			return err
		}),
	)

	m.Register(profileAwaitName, func(c *botapi.Context, sess *fsm.Session[profileState, profileData]) error {
		name := strings.TrimSpace(c.Message().Text)
		if name == "" {
			_, err := c.Reply("Имя не может быть пустым.")
			return err
		}

		sess.Data.Name = name
		if err := m.Enter(c, profileAwaitAge, sess.Data); err != nil {
			return err
		}

		_, err := c.Reply(fmt.Sprintf("Приятно познакомиться, %s! Сколько тебе лет?", name))
		return err
	})

	m.Register(profileAwaitAge, func(c *botapi.Context, sess *fsm.Session[profileState, profileData]) error {
		age, err := strconv.Atoi(strings.TrimSpace(c.Message().Text))
		if err != nil || age < 1 || age > 150 {
			_, err := c.Reply("Введи возраст числом от 1 до 150.")
			return err
		}

		if err := m.Clear(c); err != nil {
			return err
		}

		_, err = c.Reply(fmt.Sprintf(
			"Готово! %s, %d лет.\n",
			sess.Data.Name, age,
		))
		return err
	})

	pm := bot.Group(botapi.ChatTypeIs(botapi.ChatTypePrivate))

	pm.OnCommand("profile", "Заполнить профиль (демо FSM)", func(c *botapi.Context) error {
		if err := m.Enter(c, profileAwaitName, profileData{}); err != nil {
			return err
		}
		_, err := c.Reply("Как тебя зовут? /cancel — отмена.")
		return err
	})

	m.MountGroup(pm)
}

func registerCommands(bot *botapi.Bot, fsmStore fsm.Store[profileState, profileData]) {
	registerProfileFSM(bot, fsmStore)

	bot.OnCommand("start", "Welcome message and reply keyboard", func(c *botapi.Context) error {
		_, err := c.Reply("👋 Welcome! Type /help to see everything I can do.")
		return err
	}, func(u *botapi.Update) bool {
		return u.Message.Chat.Type == botapi.ChatTypePrivate
	})

	bot.Dispatcher().OnChannelParticipant(func(ctx context.Context, e tg.Entities, u *tg.UpdateChannelParticipant) error {
		if u.NewParticipant != nil {
			return nil
		}

		var chatID constant.TDLibPeerID
		chatID.Channel(u.ChannelID)

		name := "участник"
		if user, ok := e.Users[u.UserID]; ok {
			name = user.FirstName
		}

		tag := participant.Rank(u.PrevParticipant)
		if tag == "" {
			tag = name
		}

		_, err := bot.SendMessage(ctx, botapi.ID(int64(chatID)),
			fmt.Sprintf("%s покинул чат", tghtml.Mention(participant.ID(u.PrevParticipant), tag)),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	})
}
