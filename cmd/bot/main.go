package main

import (
	"activity-bot/internal/bot"
	"activity-bot/internal/config"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Config load failed: " + err.Error())
	}

	app, err := bot.NewApp(cfg)
	if err != nil {
		panic("App initialization failed: " + err.Error())
	}
	defer app.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		slog.Error("Bot execution failed", "error", err)
	}
}
