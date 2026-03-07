package service

import (
	"context"
	"io"
)

// STTProvider defines the interface for Speech-to-Text services
type STTProvider interface {
	// TranscribeStream performs real-time streaming transcription
	// Returns a channel that emits partial transcripts as they become available
	TranscribeStream(ctx context.Context, audioStream io.Reader) (<-chan TranscriptResult, error)

	// TranscribeFile performs batch transcription on a complete audio file
	TranscribeFile(ctx context.Context, audioData []byte, format string) (string, error)

	// Close releases any resources held by the provider
	Close() error
}

// TranscriptResult represents a single transcription result
type TranscriptResult struct {
	Text       string  // The transcribed text
	IsFinal    bool    // Whether this is a final result or partial
	Timestamp  int64   // Timestamp in milliseconds
	Confidence float64 // Confidence score (0-1)
	Error      error   // Error if transcription failed
}
