package repository

import (
	"database/sql"
	"fmt"

	"github.com/mino/backend/internal/model"
)

type ExtensionRepository struct {
	db *sql.DB
}

func NewExtensionRepository(db *sql.DB) *ExtensionRepository {
	return &ExtensionRepository{db: db}
}

func (r *ExtensionRepository) List(userID string) ([]*model.Extension, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, description, icon, enabled, config, created_at, updated_at
		 FROM mino_extensions WHERE user_id = $1 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exts []*model.Extension
	for rows.Next() {
		e := &model.Extension{}
		if err := rows.Scan(&e.ID, &e.UserID, &e.Name, &e.Description, &e.Icon,
			&e.Enabled, &e.Config, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		exts = append(exts, e)
	}
	return exts, rows.Err()
}

func (r *ExtensionRepository) FindByID(id, userID string) (*model.Extension, error) {
	e := &model.Extension{}
	err := r.db.QueryRow(
		`SELECT id, user_id, name, description, icon, enabled, config, created_at, updated_at
		 FROM mino_extensions WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&e.ID, &e.UserID, &e.Name, &e.Description, &e.Icon,
		&e.Enabled, &e.Config, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (r *ExtensionRepository) Create(e *model.Extension) error {
	return r.db.QueryRow(
		`INSERT INTO mino_extensions (user_id, name, description, icon, enabled, config)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		e.UserID, e.Name, e.Description, e.Icon, e.Enabled, e.Config,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *ExtensionRepository) Update(e *model.Extension) error {
	res, err := r.db.Exec(
		`UPDATE mino_extensions SET name=$1, description=$2, icon=$3, enabled=$4, config=$5, updated_at=NOW()
		 WHERE id=$6 AND user_id=$7`,
		e.Name, e.Description, e.Icon, e.Enabled, e.Config, e.ID, e.UserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("extension not found")
	}
	return nil
}

func (r *ExtensionRepository) Delete(id, userID string) error {
	res, err := r.db.Exec(`DELETE FROM mino_extensions WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("extension not found")
	}
	return nil
}
