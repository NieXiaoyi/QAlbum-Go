package model

import "time"

type File struct {
	ID          int        `json:"id" db:"id"`
	SpaceID     int        `json:"space_id" db:"space_id"`
	AlbumID     int        `json:"album_id" db:"album_id"`
	Filename    string     `json:"filename" db:"filename"`
	FileType    string     `json:"file_type" db:"file_type"`
	FileSize    int64      `json:"file_size" db:"file_size"`
	MimeType    string     `json:"mime_type" db:"mime_type"`
	StoragePath string     `json:"storage_path" db:"storage_path"`
	ThumbnailPath *string  `json:"thumbnail_path" db:"thumbnail_path"`
	CoverPath   *string    `json:"cover_path" db:"cover_path"`
	PHash       *string    `json:"phash" db:"phash"`
	Width       *int       `json:"width" db:"width"`
	Height      *int       `json:"height" db:"height"`
	Duration    *int       `json:"duration" db:"duration"`
	UploadedBy  int        `json:"uploaded_by" db:"uploaded_by"`
	UploadedAt  time.Time  `json:"uploaded_at" db:"uploaded_at"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
	BackupStatus string   `json:"backup_status" db:"backup_status"`
}
