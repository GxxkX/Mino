package repository

import (
	"database/sql"
	"encoding/json"
	"time"
)

// --------------- Models ---------------

type ChatSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChatMessage struct {
	ID        string          `json:"id"`
	SessionID string          `json:"sessionId"`
	UserID    string          `json:"userId"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Sources   json.RawMessage `json:"sources,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

// --------------- Repository ---------------

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// --------------- Session CRUD ---------------

func (r *ChatRepository) CreateSession(s *ChatSession) error {
	return r.db.QueryRow(
		`INSERT INTO mino_chat_sessions (user_id, title)
		 VALUES ($1, $2)
		 RETURNING id, created_at, updated_at`,
		s.UserID, s.Title,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *ChatRepository) ListSessions(userID string) ([]*ChatSession, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, title, created_at, updated_at
		 FROM mino_chat_sessions WHERE user_id = $1
		 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*ChatSession
	for rows.Next() {
		s := &ChatSession{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *ChatRepository) GetSession(sessionID, userID string) (*ChatSession, error) {
	s := &ChatSession{}
	err := r.db.QueryRow(
		`SELECT id, user_id, title, created_at, updated_at
		 FROM mino_chat_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ChatRepository) UpdateSessionTitle(sessionID, userID, title string) error {
	res, err := r.db.Exec(
		`UPDATE mino_chat_sessions SET title = $1, updated_at = NOW()
		 WHERE id = $2 AND user_id = $3`,
		title, sessionID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ChatRepository) DeleteSession(sessionID, userID string) error {
	res, err := r.db.Exec(
		`DELETE FROM mino_chat_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// TouchSession bumps updated_at so the session floats to the top of the list.
func (r *ChatRepository) TouchSession(sessionID string) {
	r.db.Exec(`UPDATE mino_chat_sessions SET updated_at = NOW() WHERE id = $1`, sessionID)
}

// --------------- Message CRUD ---------------

func (r *ChatRepository) ListMessages(sessionID, userID string) ([]*ChatMessage, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, user_id, role, content, sources, created_at
		 FROM mino_chat_messages
		 WHERE session_id = $1 AND user_id = $2
		 ORDER BY created_at ASC`,
		sessionID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*ChatMessage
	for rows.Next() {
		m := &ChatMessage{}
		var sources *[]byte
		if err := rows.Scan(&m.ID, &m.SessionID, &m.UserID, &m.Role, &m.Content, &sources, &m.CreatedAt); err != nil {
			return nil, err
		}
		if sources != nil {
			m.Sources = json.RawMessage(*sources)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (r *ChatRepository) CreateMessage(m *ChatMessage) error {
	var sources interface{}
	if len(m.Sources) > 0 {
		sources = m.Sources
	}
	return r.db.QueryRow(
		`INSERT INTO mino_chat_messages (session_id, user_id, role, content, sources)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		m.SessionID, m.UserID, m.Role, m.Content, sources,
	).Scan(&m.ID, &m.CreatedAt)
}
