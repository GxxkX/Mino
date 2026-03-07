package model

import "time"

type Conversation struct {
	ID            string     `db:"id" json:"id"`
	UserID        string     `db:"user_id" json:"userId"`
	Title         *string    `db:"title" json:"title,omitempty"`
	Summary       *string    `db:"summary" json:"summary,omitempty"`
	Transcript    string     `db:"transcript" json:"transcript"`
	AudioURL      *string    `db:"audio_url" json:"audioUrl,omitempty"`
	AudioDuration *int       `db:"audio_duration" json:"audioDuration,omitempty"`
	Language      string     `db:"language" json:"language"`
	Status        string     `db:"status" json:"status"`
	RecordedAt    *time.Time `db:"recorded_at" json:"recordedAt,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updatedAt"`
	Tags          []Tag      `db:"-" json:"tags,omitempty"`
}

type Tag struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"userId"`
	Name      string    `db:"name" json:"name"`
	Color     *string   `db:"color" json:"color,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}
