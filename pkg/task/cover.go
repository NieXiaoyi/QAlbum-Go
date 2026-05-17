package task

import (
	"bytes"
	"context"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/storage"
)

var coverSemaphore = make(chan struct{}, 2)

type CoverTask struct {
	fileDAO  *dao.FileDAO
	storage storage.Interface
	fileID   int
}

func NewCoverTask(fileDAO *dao.FileDAO, storage storage.Interface, fileID int) *CoverTask {
	return &CoverTask{
		fileDAO: fileDAO,
		storage: storage,
		fileID:  fileID,
	}
}

func (t *CoverTask) Execute() error {
	log.Printf("[TASK] Generating cover for file %d", t.fileID)

	coverSemaphore <- struct{}{}
	defer func() { <-coverSemaphore }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	file, err := t.fileDAO.GetByID(t.fileID)
	if err != nil {
		return err
	}

	if file.FileType != "video" {
		return nil
	}

	fullPath, err := t.storage.GetFullPath(file.StoragePath)
	if err != nil {
		return err
	}

	coverPath := filepath.Join(filepath.Dir(file.StoragePath), "cover_"+strings.TrimSuffix(filepath.Base(file.StoragePath), filepath.Ext(file.StoragePath))+".jpg")

	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", fullPath, "-ss", "00:00:01.000", "-vframes", "1", "-q:v", "2", "-f", "image2pipe", "-")
	cmd.Stderr = nil

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var buf []byte
	if buf, err = io.ReadAll(stdout); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	reader := bytes.NewReader(buf)
	if err := t.storage.Save(coverPath, reader); err != nil {
		return err
	}

	return t.fileDAO.UpdateCoverPath(t.fileID, coverPath)
}
