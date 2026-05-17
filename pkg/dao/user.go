package dao

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"qalbum-server/pkg/model"
)

type UserDAO struct {
	db *sqlx.DB
}

func NewUserDAO(db *sqlx.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) GetByID(id int) (*model.User, error) {
	var user model.User
	err := d.db.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (d *UserDAO) GetByOpenID(openid string) (*model.User, error) {
	var user model.User
	err := d.db.Get(&user, "SELECT * FROM users WHERE openid = ?", openid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by openid: %w", err)
	}
	return &user, nil
}

func (d *UserDAO) Create(openid string) (*model.User, error) {
	result, err := d.db.Exec("INSERT INTO users (openid) VALUES (?)", openid)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetByID(int(id))
}

func (d *UserDAO) List() ([]*model.User, error) {
	var users []*model.User
	err := d.db.Select(&users, "SELECT * FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
