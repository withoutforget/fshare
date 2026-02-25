package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/withoutforget/fshare/internal/server/handler/upload"
	fileservice "github.com/withoutforget/fshare/internal/service/file"
)

func Setup(e *gin.Engine, fileSvc *fileservice.FileService) {
	v1 := e.Group("/api/v1")

	upload.Register(v1, upload.NewHandler(fileSvc))
}
