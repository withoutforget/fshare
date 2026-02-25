package file

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/withoutforget/fshare/internal/infra/rustfs"
	"github.com/withoutforget/fshare/internal/model"
	filerepo "github.com/withoutforget/fshare/internal/repository/file"
)

const presignTTL = 24 * time.Hour

type FileService struct {
	repo       *filerepo.FileRepository
	s3         *rustfs.Client
	publicBase string
}

func New(repo *filerepo.FileRepository, s3 *rustfs.Client, publicBase string) *FileService {
	return &FileService{repo: repo, s3: s3, publicBase: strings.TrimRight(publicBase, "/")}
}

type UploadInput struct {
	OriginalName string
	Content      io.Reader
	Size         int64
	ContentType  string
	UploadedBy   string
}

type DownloadResult struct {
	Body        io.ReadCloser
	Size        int64
	ContentType string
	Filename    string
}

func (s *FileService) Upload(ctx context.Context, in UploadInput) (uuid.UUID, error) {
	fileID := uuid.New()
	//ext := filepath.Ext(in.OriginalName)
	s3Key := fileID.String()

	_, err := s.s3.Upload(ctx, s3Key, in.Content, in.Size, in.ContentType)
	if err != nil {
		return uuid.Nil, fmt.Errorf("file service: upload to s3: %w", err)
	}

	_, err = s.repo.AddFile(ctx, in.OriginalName, s3Key, time.Now().UTC(), in.UploadedBy)
	if err != nil {
		_ = s.s3.Delete(ctx, s3Key)
		return uuid.Nil, fmt.Errorf("file service: save metadata: %w", err)
	}

	return fileID, nil
}

// Download стримит файл из S3 напрямую через апишку.
func (s *FileService) Download(ctx context.Context, fileID uuid.UUID) (DownloadResult, error) {
	info, err := s.repo.GetFileInfo(ctx, fileID)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("file service: get file info: %w", err)
	}

	body, objInfo, err := s.s3.Download(ctx, info.S3Key)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("file service: download from s3: %w", err)
	}

	ct := objInfo.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}

	return DownloadResult{
		Body:        body,
		Size:        objInfo.Size,
		ContentType: ct,
		Filename:    info.Name,
	}, nil
}

// GetFileURL возвращает простой публичный URL без presigned параметров.
func (s *FileService) GetFileURL(ctx context.Context, fileID uuid.UUID) (*url.URL, error) {
	u, err := url.Parse(fmt.Sprintf("%s/fshare/%s", s.publicBase, fileID.String()))
	if err != nil {
		return nil, fmt.Errorf("file service: build url: %w", err)
	}
	return u, nil
}

func (s *FileService) GetFileInfo(ctx context.Context, fileID uuid.UUID) (model.File, error) {
	info, err := s.repo.GetFileInfo(ctx, fileID)
	if err != nil {
		return model.File{}, fmt.Errorf("file service: get file info: %w", err)
	}
	return info, nil
}

// presignTTL оставляем на случай если понадобится вернуть presigned
var _ = presignTTL
