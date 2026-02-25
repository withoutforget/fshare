package logger

import (
	"log/slog"
	"os"

	"github.com/withoutforget/fshare/internal/config"
)

func SetupLogger(cfg config.LoggerConfig) *slog.Logger {
	var lvl slog.Level
	_ = lvl.UnmarshalText([]byte(cfg.Level))

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
