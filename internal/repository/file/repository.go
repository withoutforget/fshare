package file

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) AddFile(
	ctx context.Context,
	filename string,
	path string,
	uploadAt time.Time,
	uploadBy string) uuid.UUID {
	return uuid.UUID{}
}

func (r *FileRepository) GetFileInfo(
	ctx context.Context,
	uuid uuid.UUID) struct{} {
	return struct{}{}
}
