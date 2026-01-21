package main

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/help"
	"activity-bot/internal/message"
	"activity-bot/internal/middleware"
	"activity-bot/internal/user"
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"

	"log"
	"os"
	"os/signal"
	"regexp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config load failed:", err)
	}

	b, err := bot.New(cfg.BotToken)
	if err != nil {
		log.Fatal("Bot init failed:", err)
	}
	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal("Pool init failed:", err)
	}

	queries := db.New(pool)

	helpH := help.NewHandler()

	msgRepo := postgres.NewMessageRepository(queries)
	chatRepo := postgres.NewChatRepository(queries, pool)
	userRepo := postgres.NewUserRepository(queries)
	msgService := message.NewService(msgRepo, userRepo, chatRepo, cfg.DefaultWeeklyNorm)
	chatService := chat.NewService(chatRepo, userRepo, cfg.DefaultWeeklyNorm)
	userService := user.NewService(userRepo)
	messageH := message.NewHandler(msgService)

	setNormRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(норма|norm|quota)\s+([0-9]+)\s*$`)
	setExemptRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(рест|rest|рэст)(?:\s+|$)(.*)$`)

	showExemptRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(рест|rest|рэст)(?:\s+.*)?$`)

	endExemptRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?-\s*(рест|rest|рэст)(?:\s+.*)?$`)

	addAdminRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(админ|admin)(?:\s+|$)(.*)$`)
	removeAdminRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?-\s*(админ|admin)(?:\s+.*)?$`)
	showAdminsRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(админы|admins)\s*$`)
	updateChatRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(обновить\s+чат|update\s+chat)\s*$`)

	showNormRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(норма|norm|quota)\s*$`)
	showReportRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(отчёт|отчет|report)\s*$`)
	showRolesRe := regexp.MustCompile(`^!(роли|roles)$`)
	setRoleRe := regexp.MustCompile(`^!(роль|setrole)`)

	chatH := chat.NewHandler(chatService, userService, chat.NewDateParser(), setNormRe, setExemptRe)
	ensureMemberExistsMW := middleware.NewEnsureMemberExists(chatRepo, userRepo, cfg.DefaultWeeklyNorm)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regexp.MustCompile("^/start"), helpH.Start)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regexp.MustCompile("^/help"), helpH.Help)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showNormRe, chatH.ShowNorm, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, setNormRe, chatH.SetNorm, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showReportRe, chatH.ShowWeeklyReport, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showExemptRe, chatH.ShowMemberExempt, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, setExemptRe, chatH.ExemptMember, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, endExemptRe, chatH.EndMemberExempt, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, addAdminRe, chatH.AddAdmin, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, removeAdminRe, chatH.RemoveAdmin, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showAdminsRe, chatH.ShowAdmins, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, updateChatRe, chatH.UpdateChat, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showRolesRe, chatH.ShowRoles, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, setRoleRe, chatH.SetRole, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeCallbackQueryData, regexp.MustCompile(`^approve:\d+:\d+$`), chatH.ApproveExemptRequest, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeCallbackQueryData, regexp.MustCompile(`^reject:\d+:\d+$`), chatH.RejectExemptRequest, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.Message != nil && (update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup")
		},
		func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message.LeftChatMember != nil {
				chatH.OnLeftMember(ctx, b, update)
				return
			}
			if update.Message.Text != "" {
				messageH.Message(ctx, b, update)
			}
		},
		ensureMemberExistsMW.Handle,
	)

	b.Start(ctx)
}
