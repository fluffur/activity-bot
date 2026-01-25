package main

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/member"
	"activity-bot/internal/command"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/exempt"
	"activity-bot/internal/help"
	msg "activity-bot/internal/message"
	"activity-bot/internal/stats"
	"activity-bot/internal/user"
	"context"
	"log/slog"
	"time"

	"log"
	"os"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatmember"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		panic("Config load failed: " + err.Error())
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	b, err := gotgbot.NewBot(cfg.BotToken, &gotgbot.BotOpts{})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}

	queries := db.New(pool)

	dp := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dp, &ext.UpdaterOpts{})

	statsRepository := postgres.NewStatsRepository(queries)
	exemptRepository := postgres.NewExemptRepository(queries, pool)
	memberRepository := postgres.NewMemberRepository(queries)
	userRepository := postgres.NewUserRepository(queries)
	chatRepository := postgres.NewChatRepository(queries)
	adminRepository := postgres.NewAdminRepository(queries)
	messageRepository := postgres.NewMessageRepository(queries)

	statsService := stats.NewService(statsRepository)
	exemptService := exempt.NewService(exemptRepository)
	chatService := chat.NewService(chatRepository, cfg.DefaultWeeklyNorm)
	userService := user.NewService(userRepository)
	memberService := member.NewService(memberRepository, chatRepository, userRepository)
	adminService := admin.NewService(adminRepository)
	messageService := msg.NewService(messageRepository)
	cb := command.NewBuilder(userService)

	helpHandler := help.NewHandler()
	statsHandler := stats.NewHandler(statsService, exemptService, memberService)
	chatHandler := chat.NewHandler(chatService, adminService)
	exemptHandler := exempt.NewHandler(exemptService, userService, adminService, exempt.NewDateParser())
	adminHandler := admin.NewHandler(adminService, userService)
	messageHandler := msg.NewHandler(messageService)
	memberHandler := member.NewHandler(memberService, userService, adminService)
	callHandler := call.NewHandler(adminService)
	dp.AddHandler(cb.New("start", helpHandler.Start))
	dp.AddHandler(cb.New("help", helpHandler.Help))
	dp.AddHandler(cb.New("stats", statsHandler.ShowStats).
		SetAliases("отчёт", "отчет").
		SetMaxArgs(1),
	)
	dp.AddHandler(cb.New("norm", chatHandler.ShowNorm).
		SetAliases("норма", "quota"),
	)

	dp.AddHandler(cb.New("norm", chatHandler.SetNorm).
		SetAliases("норма", "quota").
		SetTriggers("/", ".", "!", "+").
		AllowArgs(true).
		SetMaxArgs(1),
	)

	dp.AddHandler(cb.New("exempt", exemptHandler.Show).
		SetAliases("рест", "rest", "рэст").
		FallbackToSender(true).
		SetTriggers("/", ".", "!", "+"),
	)

	dp.AddHandler(cb.New("exempt", exemptHandler.Set).
		SetAliases("рест", "rest", "рэст").
		FallbackToSender(true).
		SetTriggers("/", ".", "!", "+").
		AllowArgs(true).
		SetMaxArgs(1),
	)

	dp.AddHandler(cb.New("-exempt", exemptHandler.End).
		FallbackToSender(true).
		SetAliases("-рест", "-rest", "-рэст").
		SetTriggers("/", ".", "!", ""),
	)

	dp.AddHandler(cb.New("admins", adminHandler.ListAdmins).
		SetAliases("админы", "админчики", "адмы", "модеры", "mods"),
	)

	dp.AddHandler(cb.New("администратор", adminHandler.AddAdmin).
		SetAliases("админ", "admin", "адм", "модер").
		SetTriggers("/", ".", "!", "+"),
	)

	dp.AddHandler(cb.New("-администратор", adminHandler.RemoveAdmin).
		SetAliases("-админ", "-admin", "-адм", "-модер", "-mod").
		SetTriggers("/", ".", "!", "+", ""),
	)

	dp.AddHandler(cb.New("обновить чат", memberHandler.UpdateMembersList).
		SetAliases("update chat", "update"),
	)

	dp.AddHandler(cb.New("роль", memberHandler.ShowRole).
		SetAliases("role", "title"),
	)

	dp.AddHandler(cb.New("роль", memberHandler.SetRole).
		SetAliases("role", "title").
		SetTriggers("/", ".", "!", "+").
		AllowArgs(true).
		SetMaxArgs(1),
	)

	dp.AddHandler(cb.New("роли", memberHandler.ListRoles).
		SetAliases("roles", "titles"),
	)

	dp.AddHandler(cb.New("call", callHandler.Call).
		SetAliases("калл", "колл").
		AllowArgs(true).
		SetMaxArgs(1),
	)

	dp.AddHandler(handlers.NewMessage(message.Text, messageHandler.Message))
	dp.AddHandler(handlers.NewMessage(message.LeftChatMember, memberHandler.OnLeftMember))
	dp.AddHandler(handlers.NewMyChatMember(
		chatmember.NewStatus("administrator"), memberHandler.OnBotPromote),
	)

	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	logger.Info("Bot has been started...", "bot_username", b.User.Username)

	updater.Idle()
}
