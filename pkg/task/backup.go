package task

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/storage"
	"qalbum-server/pkg/utils"
)

type BackupTask struct {
	fileDAO   *dao.FileDAO
	spaceDAO  *dao.SpaceDAO
	storage   storage.Interface
	fileID    int
	file      interface{}
	backupDir string
}

func NewBackupTask(fileDAO *dao.FileDAO, spaceDAO *dao.SpaceDAO, storage storage.Interface, fileID int, backupDir string) *BackupTask {
	return &BackupTask{
		fileDAO:   fileDAO,
		spaceDAO:  spaceDAO,
		storage:   storage,
		fileID:    fileID,
		backupDir: backupDir,
	}
}

func (t *BackupTask) Execute() error {
	log.Printf("[TASK] Backing up file %d", t.fileID)

	file, err := t.fileDAO.GetByID(t.fileID)
	if err != nil {
		return err
	}

	space, err := t.spaceDAO.GetByID(file.SpaceID)
	if err != nil {
		return err
	}

	if space.BackupPath == "" {
		return fmt.Errorf("no backup path configured")
	}

	reader, err := t.storage.Open(file.StoragePath)
	if err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}
	defer reader.Close()

	srcPath, err := t.storage.GetFullPath(file.StoragePath)
	if err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}

	srcMD5, err := utils.ComputeMD5(srcPath)
	if err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}

	dstPath := filepath.Join(space.BackupPath, strings.TrimPrefix(file.StoragePath, "photos"))
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}

	dstMD5, err := utils.ComputeMD5(dstPath)
	if err != nil {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return err
	}

	if srcMD5 != dstMD5 {
		t.fileDAO.UpdateBackupStatus(t.fileID, "pending")
		return fmt.Errorf("MD5 mismatch after backup")
	}

	return t.fileDAO.UpdateBackupStatus(t.fileID, "ok")
}
