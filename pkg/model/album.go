package model

import "time"

type Album struct {
	ID        int        `json:"id" db:"id"`
	SpaceID   int        `json:"space_id" db:"space_id"`
	Name      string     `json:"name" db:"name"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}
