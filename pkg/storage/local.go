package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (ls *LocalStorage) Save(path string, reader io.Reader) error {
	fullPath := filepath.Join(ls.basePath, path)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (ls *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(ls.basePath, path)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (ls *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(ls.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (ls *LocalStorage) Open(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(ls.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (ls *LocalStorage) GetFullPath(path string) (string, error) {
	return filepath.Join(ls.basePath, path), nil
}
