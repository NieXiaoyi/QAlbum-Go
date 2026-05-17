package dao

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"qalbum-server/pkg/model"
)

type SpaceDAO struct {
	db *sqlx.DB
}

func NewSpaceDAO(db *sqlx.DB) *SpaceDAO {
	return &SpaceDAO{db: db}
}

func (d *SpaceDAO) GetByID(id int) (*model.Space, error) {
	var space model.Space
	err := d.db.Get(&space, "SELECT * FROM spaces WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get space by id: %w", err)
	}
	return &space, nil
}

func (d *SpaceDAO) ListByUserID(userID int) ([]*model.Space, error) {
	var spaces []*model.Space
	query := `
		SELECT s.*, 
		       (SELECT COUNT(*) FROM space_members WHERE space_id = s.id) as member_count,
		       (SELECT COUNT(*) FROM files WHERE space_id = s.id AND backup_status = 'pending' AND deleted_at IS NULL) as pending_backup_count
		FROM spaces s
		INNER JOIN space_members sm ON s.id = sm.space_id
		WHERE sm.user_id = ?
		ORDER BY s.created_at DESC
	`
	err := d.db.Select(&spaces, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list spaces by user id: %w", err)
	}
	return spaces, nil
}

func (d *SpaceDAO) Create(name string, ownerID int, quotaBytes int64, backupPath string) (*model.Space, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec("INSERT INTO spaces (name, owner_id, quota_bytes, backup_path) VALUES (?, ?, ?, ?)",
		name, ownerID, quotaBytes, backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create space: %w", err)
	}

	spaceID, _ := result.LastInsertId()

	_, err = tx.Exec("INSERT INTO space_members (space_id, user_id, role) VALUES (?, ?, 'admin')",
		spaceID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return d.GetByID(int(spaceID))
}

func (d *SpaceDAO) Update(id int, name string, quotaBytes int64, backupPath string) (*model.Space, error) {
	_, err := d.db.Exec(`
		UPDATE spaces 
		SET name = ?, quota_bytes = ?, backup_path = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, name, quotaBytes, backupPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update space: %w", err)
	}
	return d.GetByID(id)
}

func (d *SpaceDAO) UpdateBackupPath(id int, backupPath string) error {
	_, err := d.db.Exec(`
		UPDATE spaces 
		SET backup_path = ?, backup_status = CASE WHEN ? IS NOT NULL THEN 'ok' ELSE 'unavailable' END, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, backupPath, backupPath, id)
	if err != nil {
		return fmt.Errorf("failed to update backup path: %w", err)
	}

	if backupPath != "" {
		_, err = d.db.Exec("UPDATE files SET backup_status = 'pending' WHERE space_id = ? AND deleted_at IS NULL", id)
		if err != nil {
			return fmt.Errorf("failed to mark files for rebackup: %w", err)
		}
	}

	return nil
}

func (d *SpaceDAO) Delete(id int) error {
	_, err := d.db.Exec("DELETE FROM spaces WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete space: %w", err)
	}
	return nil
}

func (d *SpaceDAO) GetMemberRole(spaceID, userID int) (string, error) {
	var role string
	err := d.db.Get(&role, "SELECT role FROM space_members WHERE space_id = ? AND user_id = ?", spaceID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get member role: %w", err)
	}
	return role, nil
}

func (d *SpaceDAO) IsMember(spaceID, userID int) (bool, error) {
	var count int
	err := d.db.Get(&count, "SELECT COUNT(*) FROM space_members WHERE space_id = ? AND user_id = ?", spaceID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	return count > 0, nil
}

func (d *SpaceDAO) AddMember(spaceID, userID int) error {
	_, err := d.db.Exec("INSERT INTO space_members (space_id, user_id, role) VALUES (?, ?, 'member')", spaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

func (d *SpaceDAO) RemoveMember(spaceID, userID int) error {
	_, err := d.db.Exec("DELETE FROM space_members WHERE space_id = ? AND user_id = ?", spaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

func (d *SpaceDAO) GetMembers(spaceID int) ([]*model.SpaceMember, error) {
	var members []*model.SpaceMember
	err := d.db.Select(&members, "SELECT user_id, role, joined_at FROM space_members WHERE space_id = ?", spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}
	return members, nil
}

func (d *SpaceDAO) CheckQuota(spaceID int) (int64, int64, error) {
	tx, err := d.db.BeginTxx(nil, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var used, quota int64
	err = tx.QueryRow(`
		SELECT 
			COALESCE(SUM(f.file_size), 0) as used,
			s.quota_bytes
		FROM spaces s
		LEFT JOIN files f ON s.id = f.space_id AND f.deleted_at IS NULL
		WHERE s.id = ?
	`, spaceID).Scan(&used, &quota)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to check quota: %w", err)
	}

	return used, quota, nil
}
