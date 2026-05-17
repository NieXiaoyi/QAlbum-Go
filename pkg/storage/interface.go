package storage

import (
	"io"
)

type Interface interface {
	Save(path string, reader io.Reader) error
	Delete(path string) error
	Exists(path string) (bool, error)
	Open(path string) (io.ReadCloser, error)
	GetFullPath(path string) (string, error)
}

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}
