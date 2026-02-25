package upload

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	fileservice "github.com/withoutforget/fshare/internal/service/file"
)

type Handler struct {
	svc *fileservice.FileService
}

func NewHandler(svc *fileservice.FileService) *Handler {
	return &Handler{svc: svc}
}

// Upload godoc
// POST /api/v1/files/
// Content-Type: multipart/form-data
// Form field: file
//
// Response 201: { "id": "<uuid>" }
func (h *Handler) Upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	id, err := h.svc.Upload(c.Request.Context(), fileservice.UploadInput{
		OriginalName: fh.Filename,
		Content:      f,
		Size:         fh.Size,
		ContentType:  fh.Header.Get("Content-Type"),
		UploadedBy:   c.ClientIP(), // TODO: заменить на user ID после добавления авторизации
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetURL godoc
// GET /api/v1/files/:id/url
//
// Response 200: { "url": "<presigned url>" }
func (h *Handler) GetURL(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	link, err := h.svc.GetFileURL(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": link.String()})
}
