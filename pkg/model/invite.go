package model

import "time"

type InviteToken struct {
	ID        int        `json:"id" db:"id"`
	SpaceID   int        `json:"space_id" db:"space_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type SpaceMember struct {
	UserID    int       `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}
