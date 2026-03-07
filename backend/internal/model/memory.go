package model

import "time"

type Memory struct {
	ID             string    `db:"id" json:"id"`
	UserID         string    `db:"user_id" json:"userId"`
	ConversationID *string   `db:"conversation_id" json:"conversationId,omitempty"`
	Content        string    `db:"content" json:"content"`
	Category       *string   `db:"category" json:"category,omitempty"`
	Importance     int       `db:"importance" json:"importance"`
	EmbeddingID    *int64    `db:"embedding_id" json:"embeddingId,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at" json:"updatedAt"`
}
