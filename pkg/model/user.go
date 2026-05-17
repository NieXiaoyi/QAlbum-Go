package model

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	OpenID    string    `json:"openid" db:"openid"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
