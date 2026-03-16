package service

import (
	"context"
	"io"

	"github.com/mino/backend/internal/model"
)

// STTProvider defines the interface for Speech-to-Text services
type STTProvider interface {
	// TranscribeStream performs real-time streaming transcription
	// Returns a channel that emits partial transcripts as they become available
	TranscribeStream(ctx context.Context, audioStream io.Reader) (<-chan TranscriptResult, error)

	// TranscribeFile performs batch transcription on a complete audio file
	TranscribeFile(ctx context.Context, audioData []byte, format string) (string, error)

	// DiarizeFile performs speaker diarization + transcription on a complete audio file.
	// Returns diarized segments with speaker labels and per-speaker embeddings.
	DiarizeFile(ctx context.Context, audioData []byte, format string) (*DiarizedResult, error)

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

// DiarizedResult holds the full result of speaker diarization + transcription.
type DiarizedResult struct {
	Text              string                     // Full transcribed text
	Language          string                     // Detected language
	Segments          []model.DiarizedSegment    // Speaker-labeled segments
	SpeakerEmbeddings map[string][]float32       // speaker_label -> embedding vector
	NumSpeakers       int                        // Number of unique speakers detected
}
