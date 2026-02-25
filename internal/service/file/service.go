package file

import "github.com/withoutforget/fshare/internal/repository/file"

type FileSerivce struct {
	repo *file.FileRepository
}

func New(repo *file.FileRepository) *FileSerivce {
	return &FileSerivce{repo: repo}
}
