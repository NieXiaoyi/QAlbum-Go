package service

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"path/filepath"
	"strings"
	"time"
	_ "image/jpeg"
	_ "image/png"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/model"
	"qalbum-server/pkg/storage"
	"qalbum-server/pkg/task"
)

type FileService struct {
	fileDAO   *dao.FileDAO
	albumDAO  *dao.AlbumDAO
	spaceDAO  *dao.SpaceDAO
	storage   storage.Interface
	taskQueue *task.Queue
}

type UploadFileMeta struct {
	FileName string
	MimeType string
	Size     int64
	Reader   io.Reader
	Data     []byte
}

func NewFileService(fileDAO *dao.FileDAO, albumDAO *dao.AlbumDAO, spaceDAO *dao.SpaceDAO, storage storage.Interface, taskQueue *task.Queue) *FileService {
	return &FileService{
		fileDAO:   fileDAO,
		albumDAO:  albumDAO,
		spaceDAO:  spaceDAO,
		storage:   storage,
		taskQueue: taskQueue,
	}
}

func (s *FileService) Upload(spaceID, albumID, userID int, meta *UploadFileMeta) (*model.File, error) {
	used, quota, err := s.spaceDAO.CheckQuota(spaceID)
	if err != nil {
		return nil, fmt.Errorf("quota check failed: %w", err)
	}

	if used + meta.Size > quota {
		return nil, fmt.Errorf("quota exceeded")
	}

	fileType := "image"
	if strings.HasPrefix(meta.MimeType, "video/") {
		fileType = "video"
	}

	suffix := make([]byte, 8)
	rand.Read(suffix)
	randomSuffix := hex.EncodeToString(suffix)

	storagePath := filepath.Join("photos", fmt.Sprintf("%d", spaceID), fmt.Sprintf("%d_%s%s", time.Now().Unix(), randomSuffix, filepath.Ext(meta.FileName)))

	width, height, duration := extractMetadata(fileType, meta.Data)

	file := &model.File{
		SpaceID:    spaceID,
		AlbumID:    albumID,
		Filename:   meta.FileName,
		FileType:   fileType,
		FileSize:   meta.Size,
		MimeType:   meta.MimeType,
		StoragePath: storagePath,
		Width:      width,
		Height:     height,
		Duration:   duration,
		UploadedBy: userID,
	}

	if err := s.storage.Save(storagePath, meta.Reader); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	createdFile, err := s.fileDAO.Create(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	if fileType == "image" {
		if s.taskQueue != nil {
			s.taskQueue.Enqueue(task.NewThumbnailTask(s.fileDAO, s.storage, createdFile.ID))
			s.taskQueue.Enqueue(task.NewPHashTask(s.fileDAO, s.storage, createdFile.ID))
		}
	} else if fileType == "video" {
		if s.taskQueue != nil {
			s.taskQueue.Enqueue(task.NewCoverTask(s.fileDAO, s.storage, createdFile.ID))
		}
	}

	space, err := s.spaceDAO.GetByID(spaceID)
	if err == nil && space.BackupPath != "" {
		if s.taskQueue != nil {
			s.taskQueue.Enqueue(task.NewBackupTask(s.fileDAO, s.spaceDAO, s.storage, createdFile.ID, space.BackupPath))
		}
	}

	return createdFile, nil
}

func extractMetadata(fileType string, data []byte) (*int, *int, *int) {
	reader := bytes.NewReader(data)

	if fileType == "image" {
		img, _, err := image.DecodeConfig(reader)
		if err == nil {
			width := img.Width
			height := img.Height
			return &width, &height, nil
		}
	}

	return nil, nil, nil
}

func (s *FileService) GetByID(id int, userID int) (*model.File, error) {
	file, err := s.fileDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	isMember, err := s.spaceDAO.IsMember(file.SpaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	return file, nil
}

func (s *FileService) ListByAlbum(spaceID, albumID, userID int, page, pageSize int) ([]*model.File, int, error) {
	isMember, err := s.spaceDAO.IsMember(spaceID, userID)
	if err != nil || !isMember {
		return nil, 0, fmt.Errorf("not a member")
	}

	return s.fileDAO.ListByAlbumPaginated(albumID, page, pageSize)
}

func (s *FileService) Update(id int, userID int, filename string, albumID int) (*model.File, error) {
	file, err := s.fileDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	isMember, err := s.spaceDAO.IsMember(file.SpaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	if filename != "" {
		if _, err := s.fileDAO.Update(id, filename); err != nil {
			return nil, err
		}
	}

	if albumID != 0 && albumID != file.AlbumID {
		if err := s.fileDAO.UpdateAlbumID(id, albumID); err != nil {
			return nil, err
		}
	}

	return s.fileDAO.GetByID(id)
}

func (s *FileService) SoftDelete(id int, userID int) error {
	file, err := s.fileDAO.GetByID(id)
	if err != nil {
		return err
	}

	isMember, err := s.spaceDAO.IsMember(file.SpaceID, userID)
	if err != nil || !isMember {
		return fmt.Errorf("not a member")
	}

	return s.fileDAO.SoftDelete(id)
}

func (s *FileService) Download(id int, userID int) (io.ReadCloser, string, error) {
	file, err := s.fileDAO.GetByID(id)
	if err != nil {
		return nil, "", err
	}

	isMember, err := s.spaceDAO.IsMember(file.SpaceID, userID)
	if err != nil || !isMember {
		return nil, "", fmt.Errorf("not a member")
	}

	reader, err := s.storage.Open(file.StoragePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	return reader, file.MimeType, nil
}

func (s *FileService) GetThumbnail(id int, userID int) (io.ReadCloser, error) {
	file, err := s.fileDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	isMember, err := s.spaceDAO.IsMember(file.SpaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	if file.ThumbnailPath != nil {
		return s.storage.Open(*file.ThumbnailPath)
	}

	if file.CoverPath != nil {
		return s.storage.Open(*file.CoverPath)
	}

	return nil, fmt.Errorf("thumbnail not ready")
}
