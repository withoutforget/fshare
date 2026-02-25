package upload

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func Register(rg *gin.RouterGroup, h *Handler) {
	g := rg.Group("/upload")

	g.POST("/", h.Upload)
}

func (h *Handler) Upload(c *gin.Context) {
	_, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

}
