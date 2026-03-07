package model

import "time"

type Task struct {
	ID             string     `db:"id" json:"id"`
	UserID         string     `db:"user_id" json:"userId"`
	ConversationID *string    `db:"conversation_id" json:"conversationId,omitempty"`
	Title          string     `db:"title" json:"title"`
	Description    *string    `db:"description" json:"description,omitempty"`
	Status         string     `db:"status" json:"status"`
	Priority       string     `db:"priority" json:"priority"`
	DueDate        *time.Time `db:"due_date" json:"dueDate,omitempty"`
	CompletedAt    *time.Time `db:"completed_at" json:"completedAt,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updatedAt"`
}
