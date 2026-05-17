package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/model"
	"qalbum-server/pkg/storage"
	"qalbum-server/pkg/task"
)

type SpaceService struct {
	spaceDAO *dao.SpaceDAO
	userDAO  *dao.UserDAO
	inviteDAO *dao.InviteDAO
	fileDAO  *dao.FileDAO
	taskQueue *task.Queue
	storage  storage.Interface
}

func NewSpaceService(spaceDAO *dao.SpaceDAO, userDAO *dao.UserDAO, inviteDAO *dao.InviteDAO, fileDAO *dao.FileDAO, taskQueue *task.Queue, storage storage.Interface) *SpaceService {
	return &SpaceService{
		spaceDAO:  spaceDAO,
		userDAO:   userDAO,
		inviteDAO: inviteDAO,
		fileDAO:   fileDAO,
		taskQueue: taskQueue,
		storage:   storage,
	}
}

func (s *SpaceService) Create(name string, ownerID int, quotaBytes int64, backupPath string) (*model.Space, error) {
	return s.spaceDAO.Create(name, ownerID, quotaBytes, backupPath)
}

func (s *SpaceService) GetByID(id int) (*model.Space, error) {
	return s.spaceDAO.GetByID(id)
}

func (s *SpaceService) ListByUser(userID int) ([]*model.Space, error) {
	return s.spaceDAO.ListByUserID(userID)
}

func (s *SpaceService) Update(id int, userID int, name string, quotaBytes int64, backupPath string) (*model.Space, error) {
	space, err := s.spaceDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	role, err := s.spaceDAO.GetMemberRole(id, userID)
	if err != nil {
		return nil, fmt.Errorf("not a member")
	}

	if role != "admin" {
		return nil, fmt.Errorf("only admin can update space")
	}

	if backupPath != "" && backupPath != space.BackupPath {
		if err := s.spaceDAO.UpdateBackupPath(id, backupPath); err != nil {
			return nil, err
		}
	}

	return s.spaceDAO.Update(id, name, quotaBytes, backupPath)
}

func (s *SpaceService) Delete(id int, userID int) error {
	role, err := s.spaceDAO.GetMemberRole(id, userID)
	if err != nil {
		return fmt.Errorf("not a member")
	}

	if role != "admin" {
		return fmt.Errorf("only admin can delete space")
	}

	return s.spaceDAO.Delete(id)
}

func (s *SpaceService) CheckQuota(spaceID int, fileSize int64) error {
	used, quota, err := s.spaceDAO.CheckQuota(spaceID)
	if err != nil {
		return err
	}

	if used+fileSize > quota {
		return fmt.Errorf("quota exceeded")
	}

	return nil
}

func (s *SpaceService) GenerateInviteToken(spaceID, userID int, expireHours int) (*model.InviteToken, error) {
	role, err := s.spaceDAO.GetMemberRole(spaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("not a member")
	}

	if role != "admin" {
		return nil, fmt.Errorf("only admin can generate invite tokens")
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(expireHours) * time.Hour)

	return s.inviteDAO.Create(spaceID, token, expiresAt)
}

func (s *SpaceService) JoinByToken(userID int, token string) (*model.Space, error) {
	invite, err := s.inviteDAO.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid invite token")
	}

	if invite.UsedAt != nil {
		return nil, fmt.Errorf("invite token already used")
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, fmt.Errorf("invite token expired")
	}

	if err := s.spaceDAO.AddMember(invite.SpaceID, userID); err != nil {
		return nil, fmt.Errorf("failed to join space: %w", err)
	}

	if err := s.inviteDAO.MarkUsed(invite.ID); err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	return s.spaceDAO.GetByID(invite.SpaceID)
}

func (s *SpaceService) GetMembers(spaceID int, userID int) ([]*model.SpaceMember, error) {
	_, err := s.spaceDAO.GetMemberRole(spaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("not a member")
	}

	return s.spaceDAO.GetMembers(spaceID)
}

func (s *SpaceService) RemoveMember(spaceID, userID, targetUserID int) error {
	role, err := s.spaceDAO.GetMemberRole(spaceID, userID)
	if err != nil {
		return fmt.Errorf("not a member")
	}

	if role != "admin" {
		return fmt.Errorf("only admin can remove members")
	}

	if userID == targetUserID {
		return fmt.Errorf("cannot remove yourself")
	}

	return s.spaceDAO.RemoveMember(spaceID, targetUserID)
}

func (s *SpaceService) IsMember(spaceID, userID int) (bool, error) {
	return s.spaceDAO.IsMember(spaceID, userID)
}

func (s *SpaceService) GetMemberRole(spaceID, userID int) (string, error) {
	return s.spaceDAO.GetMemberRole(spaceID, userID)
}

func (s *SpaceService) SyncBackup(spaceID, userID int) (int, error) {
	role, err := s.spaceDAO.GetMemberRole(spaceID, userID)
	if err != nil {
		return 0, fmt.Errorf("not a member")
	}

	if role != "admin" {
		return 0, fmt.Errorf("only admin can sync backup")
	}

	space, err := s.spaceDAO.GetByID(spaceID)
	if err != nil {
		return 0, err
	}

	if space.BackupPath == "" {
		return 0, fmt.Errorf("no backup path configured")
	}

	pendingFiles, err := s.fileDAO.GetPendingBackupFiles(spaceID)
	if err != nil {
		return 0, err
	}

	for _, file := range pendingFiles {
		if s.taskQueue != nil {
			s.taskQueue.Enqueue(task.NewBackupTask(s.fileDAO, s.spaceDAO, s.storage, file.ID, space.BackupPath))
		}
	}

	return len(pendingFiles), nil
}
