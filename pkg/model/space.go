package model

import "time"

type Space struct {
	ID                int       `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	OwnerID           int       `json:"owner_id" db:"owner_id"`
	QuotaBytes        int64     `json:"quota_bytes" db:"quota_bytes"`
	BackupPath        string    `json:"backup_path" db:"backup_path"`
	BackupStatus      string    `json:"backup_status" db:"backup_status"`
	MemberCount       int       `json:"member_count" db:"member_count"`
	PendingBackupCount int      `json:"pending_backup_count" db:"pending_backup_count"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
