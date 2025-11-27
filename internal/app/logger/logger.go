package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Level  string
	Pretty bool
}

func New(cfg Config) *slog.Logger {
	var level slog.Level

	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Pretty {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
