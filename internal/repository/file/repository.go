package file

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/withoutforget/fshare/internal/model"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

// AddFile saves file metadata and returns the generated UUID.
func (r *FileRepository) AddFile(
	ctx context.Context,
	name string,
	s3Key string,
	uploadedAt time.Time,
	uploadedBy string,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO files (name, s3_key, uploaded_at, uploaded_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, name, s3Key, uploadedAt, uploadedBy).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("file repo: add file: %w", err)
	}
	return id, nil
}

// GetFileInfo returns metadata for the given file UUID.
func (r *FileRepository) GetFileInfo(ctx context.Context, id uuid.UUID) (model.File, error) {
	var f model.File
	err := r.db.QueryRow(ctx, `
		SELECT id, name, s3_key, uploaded_at, uploaded_by
		FROM files
		WHERE id = $1
	`, id).Scan(&f.ID, &f.Name, &f.S3Key, &f.UploadedAt, &f.UploadedBy)
	if err != nil {
		return model.File{}, fmt.Errorf("file repo: get file info: %w", err)
	}
	return f, nil
}
