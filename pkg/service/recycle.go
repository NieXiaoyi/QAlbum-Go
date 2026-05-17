package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/model"
	"qalbum-server/pkg/storage"
)

type RecycleService struct {
	fileDAO  *dao.FileDAO
	albumDAO *dao.AlbumDAO
	spaceDAO *dao.SpaceDAO
	storage  storage.Interface
}

func NewRecycleService(fileDAO *dao.FileDAO, albumDAO *dao.AlbumDAO, spaceDAO *dao.SpaceDAO, storage storage.Interface) *RecycleService {
	return &RecycleService{
		fileDAO:  fileDAO,
		albumDAO: albumDAO,
		spaceDAO: spaceDAO,
		storage:  storage,
	}
}

func (s *RecycleService) List(spaceID, userID int, page, pageSize int) ([]*model.File, int, error) {
	role, err := s.spaceDAO.GetMemberRole(spaceID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("not a member")
	}

	if role != "admin" {
		return nil, 0, fmt.Errorf("admin only")
	}

	return s.fileDAO.GetRecycledFiles(spaceID, page, pageSize)
}

func (s *RecycleService) Restore(spaceID, fileID, userID int) (*model.File, error) {
	role, err := s.spaceDAO.GetMemberRole(spaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("not a member")
	}

	if role != "admin" {
		return nil, fmt.Errorf("admin only")
	}

	file, err := s.fileDAO.GetByID(fileID)
	if err != nil {
		return nil, err
	}

	albums, err := s.albumDAO.ListBySpace(file.SpaceID)
	if err != nil {
		return nil, err
	}

	targetAlbumID := -1
	for _, album := range albums {
		targetAlbumID = album.ID
		break
	}

	if targetAlbumID == -1 {
		newAlbum, err := s.albumDAO.Create(file.SpaceID, "未分类")
		if err != nil {
			return nil, fmt.Errorf("failed to create default album: %w", err)
		}
		targetAlbumID = newAlbum.ID
	}

	if err := s.fileDAO.UpdateAlbumID(fileID, targetAlbumID); err != nil {
		return nil, err
	}

	result, err := s.fileDAO.GetByID(fileID)
	if err != nil {
		return nil, err
	}

	result.DeletedAt = nil
	return result, nil
}

func (s *RecycleService) Clear(spaceID, userID int) error {
	role, err := s.spaceDAO.GetMemberRole(spaceID, userID)
	if err != nil {
		return fmt.Errorf("not a member")
	}

	if role != "admin" {
		return fmt.Errorf("admin only")
	}

	files, _, err := s.fileDAO.GetRecycledFiles(spaceID, 1, 1000)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := s.storage.Delete(file.StoragePath); err != nil {
			return fmt.Errorf("failed to delete file %d: %w", file.ID, err)
		}

		if file.ThumbnailPath != nil {
			s.storage.Delete(*file.ThumbnailPath)
		}

		if file.CoverPath != nil {
			s.storage.Delete(*file.CoverPath)
		}

		space, _ := s.spaceDAO.GetByID(file.SpaceID)
		if space != nil && space.BackupPath != "" {
			backupPath := filepath.Join(space.BackupPath, strings.TrimPrefix(file.StoragePath, "photos"))
			os.Remove(backupPath)
		}

		if err := s.fileDAO.HardDelete(file.ID); err != nil {
			return fmt.Errorf("failed to hard delete file %d: %w", file.ID, err)
		}
	}

	return nil
}
