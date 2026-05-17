package dao

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"qalbum-server/pkg/model"
)

type FileDAO struct {
	db *sqlx.DB
}

func NewFileDAO(db *sqlx.DB) *FileDAO {
	return &FileDAO{db: db}
}

func (d *FileDAO) GetByID(id int) (*model.File, error) {
	var file model.File
	err := d.db.Get(&file, "SELECT * FROM files WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file by id: %w", err)
	}
	return &file, nil
}

func (d *FileDAO) ListByAlbum(albumID int) ([]*model.File, error) {
	var files []*model.File
	err := d.db.Select(&files, "SELECT f.* FROM files f INNER JOIN albums a ON f.album_id = a.id WHERE f.album_id = ? AND f.deleted_at IS NULL AND a.deleted_at IS NULL ORDER BY f.uploaded_at DESC", albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files by album: %w", err)
	}
	return files, nil
}

func (d *FileDAO) ListByAlbumPaginated(albumID int, page, pageSize int) ([]*model.File, int, error) {
	var total int
	err := d.db.Get(&total, "SELECT COUNT(*) FROM files f INNER JOIN albums a ON f.album_id = a.id WHERE f.album_id = ? AND f.deleted_at IS NULL AND a.deleted_at IS NULL", albumID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	var files []*model.File
	offset := (page - 1) * pageSize
	err = d.db.Select(&files, "SELECT f.* FROM files f INNER JOIN albums a ON f.album_id = a.id WHERE f.album_id = ? AND f.deleted_at IS NULL AND a.deleted_at IS NULL ORDER BY f.uploaded_at DESC LIMIT ? OFFSET ?", albumID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}
	return files, total, nil
}

func (d *FileDAO) Create(file *model.File) (*model.File, error) {
	result, err := d.db.Exec(`
		INSERT INTO files (space_id, album_id, filename, file_type, file_size, mime_type, storage_path, thumbnail_path, cover_path, phash, width, height, duration, uploaded_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, file.SpaceID, file.AlbumID, file.Filename, file.FileType, file.FileSize, file.MimeType, file.StoragePath,
		file.ThumbnailPath, file.CoverPath, file.PHash, file.Width, file.Height, file.Duration, file.UploadedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetByID(int(id))
}

func (d *FileDAO) Update(id int, filename string) (*model.File, error) {
	_, err := d.db.Exec("UPDATE files SET filename = ? WHERE id = ?", filename, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}
	return d.GetByID(id)
}

func (d *FileDAO) UpdateAlbumID(id, albumID int) error {
	_, err := d.db.Exec("UPDATE files SET album_id = ? WHERE id = ?", albumID, id)
	if err != nil {
		return fmt.Errorf("failed to update file album id: %w", err)
	}
	return nil
}

func (d *FileDAO) UpdateThumbPath(id int, thumbPath string) error {
	_, err := d.db.Exec("UPDATE files SET thumbnail_path = ? WHERE id = ?", thumbPath, id)
	if err != nil {
		return fmt.Errorf("failed to update thumbnail path: %w", err)
	}
	return nil
}

func (d *FileDAO) UpdateCoverPath(id int, coverPath string) error {
	_, err := d.db.Exec("UPDATE files SET cover_path = ? WHERE id = ?", coverPath, id)
	if err != nil {
		return fmt.Errorf("failed to update cover path: %w", err)
	}
	return nil
}

func (d *FileDAO) UpdatePHash(id int, phash string) error {
	_, err := d.db.Exec("UPDATE files SET phash = ? WHERE id = ?", phash, id)
	if err != nil {
		return fmt.Errorf("failed to update phash: %w", err)
	}
	return nil
}

func (d *FileDAO) UpdateBackupStatus(id int, status string) error {
	_, err := d.db.Exec("UPDATE files SET backup_status = ? WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("failed to update backup status: %w", err)
	}
	return nil
}

func (d *FileDAO) UpdateMetadata(id int, width, height, duration *int) error {
	_, err := d.db.Exec("UPDATE files SET width = ?, height = ?, duration = ? WHERE id = ?", width, height, duration, id)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	return nil
}

func (d *FileDAO) SoftDelete(id int) error {
	_, err := d.db.Exec("UPDATE files SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to soft delete file: %w", err)
	}
	return nil
}

func (d *FileDAO) HardDelete(id int) error {
	_, err := d.db.Exec("DELETE FROM files WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to hard delete file: %w", err)
	}
	return nil
}

func (d *FileDAO) MarkAllPendingBackup(spaceID int) error {
	_, err := d.db.Exec("UPDATE files SET backup_status = 'pending' WHERE space_id = ? AND deleted_at IS NULL", spaceID)
	if err != nil {
		return fmt.Errorf("failed to mark all pending backup: %w", err)
	}
	return nil
}

func (d *FileDAO) GetExpiredRecycledFiles(days int) ([]*model.File, error) {
	var files []*model.File
	err := d.db.Select(&files, "SELECT * FROM files WHERE deleted_at IS NOT NULL AND deleted_at < datetime('now', ? || ' days')", -days)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired recycled files: %w", err)
	}
	return files, nil
}

func (d *FileDAO) GetImagesBySpace(spaceID int) ([]*model.File, error) {
	var files []*model.File
	err := d.db.Select(&files, "SELECT * FROM files WHERE space_id = ? AND file_type = 'image' AND deleted_at IS NULL", spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images by space: %w", err)
	}
	return files, nil
}

func (d *FileDAO) GetFilesWithMissingMetadata(spaceID int) ([]*model.File, error) {
	var files []*model.File
	err := d.db.Select(&files, `
		SELECT * FROM files 
		WHERE space_id = ? 
			AND deleted_at IS NULL 
			AND (thumbnail_path IS NULL OR (file_type = 'video' AND cover_path IS NULL) OR phash IS NULL)
	`, spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files with missing metadata: %w", err)
	}
	return files, nil
}

func (d *FileDAO) GetPendingBackupFiles(spaceID int) ([]*model.File, error) {
	var files []*model.File
	err := d.db.Select(&files, "SELECT * FROM files WHERE space_id = ? AND backup_status = 'pending' AND deleted_at IS NULL", spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending backup files: %w", err)
	}
	return files, nil
}

func (d *FileDAO) GetRecycledFiles(spaceID int, page, pageSize int) ([]*model.File, int, error) {
	var total int
	err := d.db.Get(&total, "SELECT COUNT(*) FROM files WHERE space_id = ? AND deleted_at IS NOT NULL", spaceID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count recycled files: %w", err)
	}

	var files []*model.File
	offset := (page - 1) * pageSize
	err = d.db.Select(&files, "SELECT * FROM files WHERE space_id = ? AND deleted_at IS NOT NULL ORDER BY deleted_at DESC LIMIT ? OFFSET ?", spaceID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list recycled files: %w", err)
	}
	return files, total, nil
}

func (d *FileDAO) GetSpacesWithMissingMetadata(graceHours int) ([]int, error) {
	var spaceIDs []int
	err := d.db.Select(&spaceIDs, `
		SELECT DISTINCT space_id 
		FROM files 
		WHERE deleted_at IS NULL 
			AND uploaded_at < datetime('now', ? || ' hours')
			AND (thumbnail_path IS NULL OR (file_type = 'video' AND cover_path IS NULL) OR phash IS NULL)
	`, -graceHours)
	if err != nil {
		return nil, fmt.Errorf("failed to get spaces with missing metadata: %w", err)
	}
	return spaceIDs, nil
}
