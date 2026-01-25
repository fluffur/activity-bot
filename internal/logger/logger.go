package logger

import (
	"log/slog"
	"os"
)

var L *slog.Logger

func Init(debug bool) {
	var handler slog.Handler
	if debug {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	L = slog.New(handler)
	slog.SetDefault(L)
}
