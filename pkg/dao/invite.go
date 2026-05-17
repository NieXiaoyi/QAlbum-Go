package dao

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"qalbum-server/pkg/model"
)

type InviteDAO struct {
	db *sqlx.DB
}

func NewInviteDAO(db *sqlx.DB) *InviteDAO {
	return &InviteDAO{db: db}
}

func (d *InviteDAO) Create(spaceID int, token string, expiresAt time.Time) (*model.InviteToken, error) {
	result, err := d.db.Exec("INSERT INTO invite_tokens (space_id, token, expires_at) VALUES (?, ?, ?)", spaceID, token, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create invite token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetByID(int(id))
}

func (d *InviteDAO) GetByID(id int) (*model.InviteToken, error) {
	var token model.InviteToken
	err := d.db.Get(&token, "SELECT * FROM invite_tokens WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite token by id: %w", err)
	}
	return &token, nil
}

func (d *InviteDAO) GetByToken(token string) (*model.InviteToken, error) {
	var invite model.InviteToken
	err := d.db.Get(&invite, "SELECT * FROM invite_tokens WHERE token = ?", token)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite token by token: %w", err)
	}
	return &invite, nil
}

func (d *InviteDAO) MarkUsed(id int) error {
	_, err := d.db.Exec("UPDATE invite_tokens SET used_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to mark invite token as used: %w", err)
	}
	return nil
}

func (d *InviteDAO) DeleteExpired() error {
	_, err := d.db.Exec("DELETE FROM invite_tokens WHERE expires_at < CURRENT_TIMESTAMP AND used_at IS NULL")
	if err != nil {
		return fmt.Errorf("failed to delete expired invite tokens: %w", err)
	}
	return nil
}
