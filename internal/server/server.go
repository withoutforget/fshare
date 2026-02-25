package server

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/withoutforget/fshare/internal/config"
	"github.com/withoutforget/fshare/internal/server/handler"
	"github.com/withoutforget/fshare/internal/service/file"
)

type Server struct {
	ctx context.Context
	cfg config.Config
	e   *gin.Engine
}

func NewServer(ctx context.Context, cfg config.Config, fileSerice *file.FileService) Server {
	var s Server
	s.ctx = ctx
	s.cfg = cfg
	s.e = gin.New()
	s.e.Use(gin.Recovery())
	s.e.Use(LoggerMiddleware())
	s.e.RedirectTrailingSlash = false
	handler.Setup(s.e, fileSerice)

	return s
}

func (s Server) Run() {
	slog.Info("Starting server...")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: s.e.Handler(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	<-s.ctx.Done()
	slog.Info("Shutting down...")
	if err := srv.Shutdown(s.ctx); err != nil {
		panic("Error while shutdown")
	}
}
