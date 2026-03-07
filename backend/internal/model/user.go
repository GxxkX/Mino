package model

import "time"

type User struct {
	ID           string    `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Email        *string   `db:"email" json:"email,omitempty"`
	DisplayName  *string   `db:"display_name" json:"displayName,omitempty"`
	AvatarURL    *string   `db:"avatar_url" json:"avatarUrl,omitempty"`
	Role         string    `db:"role" json:"role"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}
