package model

import "time"

type Extension struct {
	ID          string     `db:"id" json:"id"`
	UserID      string     `db:"user_id" json:"userId"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Icon        string     `db:"icon" json:"icon"`
	Enabled     bool       `db:"enabled" json:"enabled"`
	Config      *string    `db:"config" json:"config,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
}
