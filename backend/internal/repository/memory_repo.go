package repository

import (
	"database/sql"
	"fmt"

	"github.com/mino/backend/internal/model"
)

type MemoryRepository struct {
	db *sql.DB
}

func NewMemoryRepository(db *sql.DB) *MemoryRepository {
	return &MemoryRepository{db: db}
}

func (r *MemoryRepository) List(userID string, limit, offset int) ([]*model.Memory, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM mino_memories WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		`SELECT id, user_id, conversation_id, content, category, importance, embedding_id, created_at, updated_at
		 FROM mino_memories WHERE user_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var mems []*model.Memory
	for rows.Next() {
		m := &model.Memory{}
		if err := rows.Scan(&m.ID, &m.UserID, &m.ConversationID, &m.Content,
			&m.Category, &m.Importance, &m.EmbeddingID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		mems = append(mems, m)
	}
	return mems, total, rows.Err()
}

func (r *MemoryRepository) FindByID(id, userID string) (*model.Memory, error) {
	m := &model.Memory{}
	err := r.db.QueryRow(
		`SELECT id, user_id, conversation_id, content, category, importance, embedding_id, created_at, updated_at
		 FROM mino_memories WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&m.ID, &m.UserID, &m.ConversationID, &m.Content,
		&m.Category, &m.Importance, &m.EmbeddingID, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (r *MemoryRepository) Create(m *model.Memory) error {
	return r.db.QueryRow(
		`INSERT INTO mino_memories (user_id, conversation_id, content, category, importance)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		m.UserID, m.ConversationID, m.Content, m.Category, m.Importance,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *MemoryRepository) Update(m *model.Memory) error {
	res, err := r.db.Exec(
		`UPDATE mino_memories SET content=$1, category=$2, importance=$3, updated_at=NOW()
		 WHERE id=$4 AND user_id=$5`,
		m.Content, m.Category, m.Importance, m.ID, m.UserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("memory not found")
	}
	return nil
}

func (r *MemoryRepository) Delete(id, userID string) error {
	res, err := r.db.Exec(`DELETE FROM mino_memories WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("memory not found")
	}
	return nil
}

func (r *MemoryRepository) CreateBatch(mems []*model.Memory) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO mino_memories (user_id, conversation_id, content, category, importance)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range mems {
		if err := stmt.QueryRow(m.UserID, m.ConversationID, m.Content, m.Category, m.Importance).
			Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}
