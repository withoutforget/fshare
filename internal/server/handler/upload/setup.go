package upload

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, h *Handler) {
	g := rg.Group("/files")

	g.POST("/", h.Upload)
	g.GET("/:id/url", h.GetURL)
}
