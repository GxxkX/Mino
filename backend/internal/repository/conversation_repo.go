package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mino/backend/internal/model"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) List(userID string, limit, offset int) ([]*model.Conversation, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM mino_conversations WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		`SELECT id, user_id, title, summary, transcript, audio_url, audio_duration, language, status, recorded_at, created_at, updated_at
		 FROM mino_conversations WHERE user_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var convs []*model.Conversation
	for rows.Next() {
		c := &model.Conversation{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Summary, &c.Transcript,
			&c.AudioURL, &c.AudioDuration, &c.Language, &c.Status, &c.RecordedAt,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		convs = append(convs, c)
	}
	return convs, total, rows.Err()
}

func (r *ConversationRepository) FindByID(id, userID string) (*model.Conversation, error) {
	c := &model.Conversation{}
	err := r.db.QueryRow(
		`SELECT id, user_id, title, summary, transcript, audio_url, audio_duration, language, status, recorded_at, created_at, updated_at
		 FROM mino_conversations WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.Summary, &c.Transcript,
		&c.AudioURL, &c.AudioDuration, &c.Language, &c.Status, &c.RecordedAt,
		&c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *ConversationRepository) Create(c *model.Conversation) error {
	return r.db.QueryRow(
		`INSERT INTO mino_conversations (user_id, title, summary, transcript, audio_url, audio_duration, language, status, recorded_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		c.UserID, c.Title, c.Summary, c.Transcript, c.AudioURL, c.AudioDuration,
		c.Language, c.Status, c.RecordedAt,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *ConversationRepository) Update(c *model.Conversation) error {
	res, err := r.db.Exec(
		`UPDATE mino_conversations SET title=$1, summary=$2, transcript=$3, audio_url=$4, status=$5, updated_at=NOW()
		 WHERE id=$6 AND user_id=$7`,
		c.Title, c.Summary, c.Transcript, c.AudioURL, c.Status, c.ID, c.UserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("conversation not found")
	}
	return nil
}

func (r *ConversationRepository) Delete(id, userID string) error {
	res, err := r.db.Exec(`DELETE FROM mino_conversations WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("conversation not found")
	}
	return nil
}

// Search does a simple full-text search on title + transcript
func (r *ConversationRepository) Search(userID, query string, limit int) ([]*model.Conversation, error) {
	q := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.Query(
		`SELECT id, user_id, title, summary, transcript, audio_url, audio_duration, language, status, recorded_at, created_at, updated_at
		 FROM mino_conversations
		 WHERE user_id = $1 AND (LOWER(title) LIKE $2 OR LOWER(transcript) LIKE $2)
		 ORDER BY created_at DESC LIMIT $3`,
		userID, q, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []*model.Conversation
	for rows.Next() {
		c := &model.Conversation{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Summary, &c.Transcript,
			&c.AudioURL, &c.AudioDuration, &c.Language, &c.Status, &c.RecordedAt,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, c)
	}
	return convs, rows.Err()
}
