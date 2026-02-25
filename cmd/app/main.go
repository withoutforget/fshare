package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/withoutforget/fshare/internal/config"
	"github.com/withoutforget/fshare/internal/infra/postgres"
	"github.com/withoutforget/fshare/internal/infra/rustfs"
	"github.com/withoutforget/fshare/internal/logger"
	"github.com/withoutforget/fshare/internal/repository/file"
	"github.com/withoutforget/fshare/internal/server"
	fileService "github.com/withoutforget/fshare/internal/service/file"
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

	pg, err := postgres.NewPostgres(cfg.Postgres)
	if err != nil {
		slog.Error("Error while connecting pg", slog.String("error", err.Error()))
	}

	fileRepo := file.NewFileRepository(pg.Pool)
	s3, err := rustfs.NewClient(cfg.S3)

	service := fileService.New(fileRepo, s3)

	if err != nil {
		panic(err.Error())
	}

	srv := server.NewServer(ctx, cfg, service)

	srv.Run()
}
