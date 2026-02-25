package model

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID         uuid.UUID
	Name       string
	S3Key      string
	UploadedAt time.Time
	UploadedBy string
}
