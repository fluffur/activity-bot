package main

import (
	"activity-bot/internal/bot"
	"activity-bot/internal/config"
	"log/slog"
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

	if err := app.Run(); err != nil {
		slog.Error("Bot execution failed", "error", err)
	}
}
