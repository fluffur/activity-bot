package main

import (
	"activity-bot/internal/config"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gotd/botapi"
	"github.com/gotd/botapi/storage"
	"github.com/gotd/log/logzap"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
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

	bot.Use(botapi.Recover(), botapi.Timeout(time.Minute), botapi.Logging())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	registerCommands(bot)

	log.Info("Starting bot")
	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}

// registerCommands wires the slash commands. Each description is published to
// the client command menu via SetMyCommands on start.
func registerCommands(bot *botapi.Bot) {
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

		tag := participantRank(u.PrevParticipant)
		if tag == "" {
			tag = name
		}

		_, err := bot.SendMessage(ctx, botapi.ID(int64(chatID)),
			fmt.Sprintf("%s покинул чат", mention(participantID(u.PrevParticipant), tag)),
			botapi.WithParseMode(botapi.ParseModeHTML),
		)

		return err
	})

}

func participantRank(p tg.ChannelParticipantClass) string {
	if p == nil {
		return ""
	}
	switch v := p.(type) {
	case *tg.ChannelParticipant:
		return v.Rank
	case *tg.ChannelParticipantAdmin:
		return v.Rank
	case *tg.ChannelParticipantCreator:
		return v.Rank
	case *tg.ChannelParticipantSelf:
		if rank, ok := v.GetRank(); ok {
			return rank
		}
	}
	return ""
}

func participantID(p tg.ChannelParticipantClass) int64 {
	if p == nil {
		return 0
	}
	switch v := p.(type) {
	case *tg.ChannelParticipant:
		return v.UserID
	case *tg.ChannelParticipantAdmin:
		return v.UserID
	case *tg.ChannelParticipantCreator:
		return v.UserID
	case *tg.ChannelParticipantSelf:
		return v.UserID
	}
	return 0
}

func mention(userID int64, text string) string {
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, userID, text)
}

func Command(botUsername, name string) botapi.Predicate {
	return func(u *botapi.Update) bool {
		m := u.EffectiveMessage()
		if m == nil {
			return false
		}

		got, target, ok := commandName(m.Text)
		if !ok || got != name {
			return false
		}

		return target == "" || strings.EqualFold(target, botUsername)
	}
}

// commandName extracts the bot command name and its optional @target from
// message text: "/start@bot foo" yields ("start", "bot", true), "/start foo"
// yields ("start", "", true). Pure.
func commandName(text string) (name, target string, ok bool) {
	if !strings.HasPrefix(text, "/") {
		return "", "", false
	}

	field := text
	if i := strings.IndexAny(text, " \t\n"); i >= 0 {
		field = text[:i]
	}

	field = field[1:] // drop leading slash
	if at := strings.IndexByte(field, '@'); at >= 0 {
		target = field[at+1:]
		field = field[:at]
	}

	return field, target, field != ""
}
