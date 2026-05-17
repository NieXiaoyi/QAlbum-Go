package dao

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"qalbum-server/pkg/model"
)

type AlbumDAO struct {
	db *sqlx.DB
}

func NewAlbumDAO(db *sqlx.DB) *AlbumDAO {
	return &AlbumDAO{db: db}
}

func (d *AlbumDAO) GetByID(id int) (*model.Album, error) {
	var album model.Album
	err := d.db.Get(&album, "SELECT * FROM albums WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get album by id: %w", err)
	}
	return &album, nil
}

func (d *AlbumDAO) ListBySpace(spaceID int) ([]*model.Album, error) {
	var albums []*model.Album
	err := d.db.Select(&albums, "SELECT * FROM albums WHERE space_id = ? AND deleted_at IS NULL ORDER BY created_at DESC", spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list albums by space: %w", err)
	}
	return albums, nil
}

func (d *AlbumDAO) Create(spaceID int, name string) (*model.Album, error) {
	result, err := d.db.Exec("INSERT INTO albums (space_id, name) VALUES (?, ?)", spaceID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create album: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetByID(int(id))
}

func (d *AlbumDAO) Update(id int, name string) (*model.Album, error) {
	_, err := d.db.Exec("UPDATE albums SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", name, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update album: %w", err)
	}
	return d.GetByID(id)
}

func (d *AlbumDAO) SoftDelete(id int) error {
	_, err := d.db.Exec("UPDATE albums SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to soft delete album: %w", err)
	}
	return nil
}

func (d *AlbumDAO) HardDelete(id int) error {
	_, err := d.db.Exec("DELETE FROM albums WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to hard delete album: %w", err)
	}
	return nil
}

func (d *AlbumDAO) GetExpiredDeletedAlbums(days int) ([]*model.Album, error) {
	var albums []*model.Album
	err := d.db.Select(&albums, "SELECT * FROM albums WHERE deleted_at IS NOT NULL AND deleted_at < datetime('now', ? || ' days')", -days)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired deleted albums: %w", err)
	}
	return albums, nil
}

func (d *AlbumDAO) GetSpaceByAlbumID(albumID int) (*model.Space, error) {
	var space model.Space
	err := d.db.Get(&space, `
		SELECT s.* FROM spaces s
		INNER JOIN albums a ON s.id = a.space_id
		WHERE a.id = ?
	`, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space by album id: %w", err)
	}
	return &space, nil
}
