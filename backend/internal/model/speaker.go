package model

import "time"

// SpeakerProfile represents a known speaker voice profile for a user.
type SpeakerProfile struct {
	ID              string    `json:"id" db:"id"`
	UserID          string    `json:"user_id" db:"user_id"`
	Name            string    `json:"name" db:"name"`
	MilvusSpeakerID string    `json:"milvus_speaker_id" db:"milvus_speaker_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// DiarizedSegment represents a single segment of speaker-diarized transcript.
type DiarizedSegment struct {
	Speaker     string  `json:"speaker"`
	SpeakerName string  `json:"speaker_name,omitempty"`
	Text        string  `json:"text"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
}
