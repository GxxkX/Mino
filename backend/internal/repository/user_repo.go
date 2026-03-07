package repository

import (
	"database/sql"
	"fmt"

	"github.com/mino/backend/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

type UserCreate struct {
	Username     string
	PasswordHash string
	Email        *string
	DisplayName  *string
	Role         string
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, email, display_name, avatar_url, role, created_at, updated_at
		 FROM mino_users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, email, display_name, avatar_url, role, created_at, updated_at
		 FROM mino_users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) Create(u *model.User) error {
	return r.db.QueryRow(
		`INSERT INTO mino_users (username, password_hash, email, display_name, avatar_url, role)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		u.Username, u.PasswordHash, u.Email, u.DisplayName, u.AvatarURL, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) CreateFromInput(uc *UserCreate) error {
	var id string
	return r.db.QueryRow(
		`INSERT INTO mino_users (username, password_hash, email, display_name, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		uc.Username, uc.PasswordHash, uc.Email, uc.DisplayName, uc.Role,
	).Scan(&id)
}

func (r *UserRepository) UpdatePassword(id, passwordHash string) error {
	res, err := r.db.Exec(
		`UPDATE mino_users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		passwordHash, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
