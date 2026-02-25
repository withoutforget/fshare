package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/withoutforget/fshare/internal/config"
	"github.com/withoutforget/fshare/internal/logger"
	"github.com/withoutforget/fshare/internal/server"
)

func main() {
	filename := os.Getenv("CONFIG_PATH")
	if filename == "" {
		filename = "./config/config.toml"
	}
	cfg := config.NewConfig(filename)

	logger := logger.SetupLogger(cfg.Logger)
	logger.Info("Config loaded, logger set up.")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGABRT)
	defer cancel()
	srv := server.NewServer(ctx, cfg)

	srv.Run()
}
