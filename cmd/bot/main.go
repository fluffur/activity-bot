package main

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/member"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/exempt"
	"activity-bot/internal/help"
	"activity-bot/internal/helpers"
	"activity-bot/internal/message"
	"activity-bot/internal/middleware"
	"activity-bot/internal/stats"
	"activity-bot/internal/user"
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"

	"fmt"
	"log"
	"net/http"
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

	adminRepo := postgres.NewAdminRepository(queries)
	chatRepo := postgres.NewChatRepository(queries)
	exemptRepo := postgres.NewExemptRepository(queries, pool)
	memberRepo := postgres.NewMemberRepository(queries)
	msgRepo := postgres.NewMessageRepository(queries)
	statsRepo := postgres.NewStatsRepository(queries)
	userRepo := postgres.NewUserRepository(queries)

	adminService := admin.NewService(adminRepo)
	chatService := chat.NewService(chatRepo, cfg.DefaultWeeklyNorm)
	exemptService := exempt.NewService(exemptRepo)
	memberService := member.NewService(memberRepo, userRepo)
	msgService := message.NewService(msgRepo)
	statsService := stats.NewService(statsRepo)
	userService := user.NewService(userRepo)

	setNormRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(норма|norm|quota)\s+([0-9]+)\s*$`)
	setExemptRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(рест|rest|рэст)(?:\s+|$)(.*)$`)
	showExemptRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(рест|rest|рэст)(?:\s+.*)?$`)
	endExemptRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?-\s*(рест|rest|рэст)(?:\s+.*)?$`)
	addAdminRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(админ|admin)(?:\s+|$)(.*)$`)
	removeAdminRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?-\s*(админ|admin)(?:\s+.*)?$`)
	showAdminsRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(админы|admins|адмы)(?:\s+.*)?$`)
	updateChatRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(обновить\s+чат|update\s+chat)\s*$`)
	showNormRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(норма|norm|quota)\s*$`)
	showReportRe := regexp.MustCompile(`(?i)^(?:[!/.]\s*)?(отчёт|отчет|stats)\s*$`)
	showRolesRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(роли|roles)\s*$`)
	setRoleRe := regexp.MustCompile(`(?i)^(?:[!/.+]\s*)?(роль|setrole)(?:\s+|$)`)

	adminHandler := admin.NewHandler(adminService, userService)
	chatHandler := chat.NewHandler(chatService, adminService, setNormRe)
	exemptHandler := exempt.NewHandler(exemptService, userService, adminService, exempt.NewDateParser(), setExemptRe)
	memberHandler := member.NewHandler(memberService, userService, adminService, setRoleRe)
	statsHandler := stats.NewHandler(statsService, exemptService, memberService)

	helpHandler := help.NewHandler()

	messageH := message.NewHandler(msgService)
	ensureMemberExistsMW := middleware.NewEnsureMemberExists(chatService, userService, memberService)

	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommandStartOnly, helpHandler.Start)
	b.RegisterHandler(bot.HandlerTypeMessageText, "help", bot.MatchTypeCommandStartOnly, helpHandler.Help)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showNormRe, chatHandler.ShowNorm)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, setNormRe, chatHandler.SetNorm, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showReportRe, statsHandler.ShowWeeklyReport, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, setExemptRe, exemptHandler.ExemptMember, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showExemptRe, exemptHandler.ShowMemberExempt, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, endExemptRe, exemptHandler.EndMemberExempt, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showAdminsRe, adminHandler.ListAdmins, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, addAdminRe, adminHandler.AddAdmin, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, removeAdminRe, adminHandler.RemoveAdmin, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, updateChatRe, memberHandler.UpdateMembersList, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, showRolesRe, memberHandler.ListRoles, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, setRoleRe, memberHandler.SetRole, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerRegexp(bot.HandlerTypeCallbackQueryData, regexp.MustCompile(`^approve:\d+:\d+$`), exemptHandler.ApproveExemptRequest, middleware.OnlyGroups, ensureMemberExistsMW.Handle)
	b.RegisterHandlerRegexp(bot.HandlerTypeCallbackQueryData, regexp.MustCompile(`^reject:\d+:\d+$`), exemptHandler.RejectExemptRequest, middleware.OnlyGroups, ensureMemberExistsMW.Handle)

	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.Message != nil && (update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup")
		},
		func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message.LeftChatMember != nil {
				memberHandler.OnLeftMember(ctx, b, update)
				return
			}
			if update.Message.Text != "" {
				messageH.Message(ctx, b, update)
			}
		},
		ensureMemberExistsMW.Handle,
	)

	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.MyChatMember != nil
		},
		func(ctx context.Context, b *bot.Bot, update *models.Update) {
			chatID := update.MyChatMember.Chat.ID

			if update.MyChatMember.NewChatMember.Administrator != nil {
				count, err := helpers.UpdateChatMembers(ctx, b, memberService, chatID)
				if err != nil {
					log.Println("Failed to update chat members on join:", err)
					return
				}
				log.Printf("Updated chat %d members on bot join, total %d members\n", chatID, count)
			}
		},
	)
	commands := []models.BotCommand{
		{Command: "help", Description: "Показать список всех команд"},
		{Command: "stats", Description: "Получить еженедельный отчёт по активности"},
		{Command: "norm", Description: "Показать текущую норму сообщений"},
		{Command: "rest", Description: "Показать или поставить рест"},
		{Command: "role", Description: "Показать или изменить роль пользователя"},
		{Command: "roles", Description: "Список всех участников с ролями"},
		{Command: "admins", Description: "Список администраторов бота"},
	}

	if _, err := b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: commands,
	}); err != nil {
		log.Println("Failed to set bot commands:", err)
	} else {
		log.Println("Bot commands set successfully")
	}

	if cfg.WebhookURL != "" {
		log.Printf("Setting up webhook at %s", cfg.WebhookURL+"/telegram/webhook")

		_, err = b.SetWebhook(ctx, &bot.SetWebhookParams{
			URL:         cfg.WebhookURL + "/telegram/webhook",
			SecretToken: cfg.WebhookSecretToken,
		})
		if err != nil {
			log.Fatal("SetWebhook failed:", err)
		}

		go func() {
			addr := fmt.Sprintf(":%d", cfg.HTTPPort)
			mux := http.NewServeMux()
			mux.Handle("/telegram/webhook", b.WebhookHandler())

			log.Printf("Starting webhook server on %s", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Fatal("Webhook server failed:", err)
			}
		}()
	} else {
		log.Println("Starting bot in long polling mode")
		go b.Start(ctx)
	}

	<-ctx.Done()
}
