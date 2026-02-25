package upload

import (
	"fmt"
	"io"
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
		UploadedBy:   c.ClientIP(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

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

// Download стримит файл прямо в ответ.
// GET /api/v1/files/:id
func (h *Handler) Download(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	result, err := h.svc.Download(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer result.Body.Close()

	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", result.Size))
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, result.Filename))
	c.Header("Cache-Control", "public, max-age=86400")

	c.Status(http.StatusOK)
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 32*1024)
		n, err := result.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		return err == nil
	})
}
