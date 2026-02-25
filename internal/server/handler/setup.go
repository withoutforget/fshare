package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/withoutforget/fshare/internal/server/handler/upload"
)

func Setup(e *gin.Engine) {
	v1 := e.Group("/api/v1")

	upload.Register(v1, upload.NewHandler())
}
