package logging

import (
	"log/slog"
	"os"
)

func New(level slog.Level, format string) *slog.Logger {
	options := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	return slog.New(handler)
}
