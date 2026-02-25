package file

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/withoutforget/fshare/internal/infra/rustfs"
	"github.com/withoutforget/fshare/internal/model"
	filerepo "github.com/withoutforget/fshare/internal/repository/file"
)

const presignTTL = 24 * time.Hour

type FileService struct {
	repo *filerepo.FileRepository
	s3   *rustfs.Client
}

func New(repo *filerepo.FileRepository, s3 *rustfs.Client) *FileService {
	return &FileService{repo: repo, s3: s3}
}

// UploadInput carries everything needed to store one file.
type UploadInput struct {
	// OriginalName is the original filename from the client (e.g. "report.pdf").
	OriginalName string
	// Content is the file body. The caller is responsible for closing it.
	Content io.Reader
	// Size is the exact content length in bytes, or -1 if unknown.
	Size int64
	// ContentType is the MIME type (e.g. "image/jpeg").
	// Pass "" to let the storage layer detect it.
	ContentType string
	// UploadedBy is any identifier of the uploader (user ID, email, etc.).
	UploadedBy string
}

// Upload stores the file in S3 and saves metadata to the database.
// Returns the newly assigned file UUID that can be used to retrieve the file later.
func (s *FileService) Upload(ctx context.Context, in UploadInput) (uuid.UUID, error) {
	// Build a unique S3 key: <uuid>/<original_filename>
	// Keeping the original name inside the key makes presigned URLs
	// show a meaningful filename in the browser download dialog.
	fileID := uuid.New()
	ext := filepath.Ext(in.OriginalName)
	s3Key := fmt.Sprintf("%s%s", fileID.String(), ext)

	// 1. Push bytes to object storage first.
	// If the DB insert fails afterwards we end up with an orphan object —
	// acceptable for now; a background cleanup job can handle it later.
	_, err := s.s3.Upload(ctx, s3Key, in.Content, in.Size, in.ContentType)
	if err != nil {
		return uuid.Nil, fmt.Errorf("file service: upload to s3: %w", err)
	}

	// 2. Persist metadata.
	id, err := s.repo.AddFile(ctx, in.OriginalName, s3Key, time.Now().UTC(), in.UploadedBy)
	if err != nil {
		// Best-effort cleanup: try to remove the orphan object.
		_ = s.s3.Delete(ctx, s3Key)
		return uuid.Nil, fmt.Errorf("file service: save metadata: %w", err)
	}

	return id, nil
}

// GetFileURL returns a short-lived presigned URL for downloading the file.
// The URL is valid for 24 hours.
func (s *FileService) GetFileURL(ctx context.Context, fileID uuid.UUID) (*url.URL, error) {
	// 1. Fetch metadata to get the S3 key.
	info, err := s.repo.GetFileInfo(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file service: get file info: %w", err)
	}

	// 2. Generate a presigned GET URL.
	link, err := s.s3.PresignedGetURL(ctx, info.S3Key, presignTTL)
	if err != nil {
		return nil, fmt.Errorf("file service: presign url: %w", err)
	}

	return link, nil
}

// GetFileInfo returns the stored metadata for a file without generating a URL.
func (s *FileService) GetFileInfo(ctx context.Context, fileID uuid.UUID) (model.File, error) {
	info, err := s.repo.GetFileInfo(ctx, fileID)
	if err != nil {
		return model.File{}, fmt.Errorf("file service: get file info: %w", err)
	}
	return info, nil
}
