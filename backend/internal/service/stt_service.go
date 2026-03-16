package service

import (
	"context"
	"fmt"
	"io"

	"github.com/mino/backend/internal/config"
)

// STTService manages STT providers and provides a unified interface
type STTService struct {
	provider STTProvider
	config   *config.STTConfig
}

// NewSTTService creates a new STT service with the specified provider
func NewSTTService(cfg *config.STTConfig) (*STTService, error) {
	var provider STTProvider

	switch cfg.Provider {
	case "whisper":
		provider = NewWhisperSTT(
			cfg.WhisperAPIURL,
			cfg.WhisperAPIKey,
			cfg.WhisperModel,
			cfg.WhisperLanguage,
		)

	default:
		return nil, fmt.Errorf("unsupported STT provider: %s (supported: whisper)", cfg.Provider)
	}

	return &STTService{
		provider: provider,
		config:   cfg,
	}, nil
}

// TranscribeStream performs real-time streaming transcription
func (s *STTService) TranscribeStream(ctx context.Context, audioStream io.Reader) (<-chan TranscriptResult, error) {
	return s.provider.TranscribeStream(ctx, audioStream)
}

// TranscribeFile performs batch transcription on a complete audio file
func (s *STTService) TranscribeFile(ctx context.Context, audioData []byte, format string) (string, error) {
	return s.provider.TranscribeFile(ctx, audioData, format)
}

// DiarizeFile performs speaker diarization + transcription on a complete audio file
func (s *STTService) DiarizeFile(ctx context.Context, audioData []byte, format string) (*DiarizedResult, error) {
	return s.provider.DiarizeFile(ctx, audioData, format)
}

// IsDiarizationEnabled returns whether pyannote diarization is configured
func (s *STTService) IsDiarizationEnabled() bool {
	return s.config.PyannoteEnabled
}

// Close releases any resources held by the service
func (s *STTService) Close() error {
	if s.provider != nil {
		return s.provider.Close()
	}
	return nil
}

// GetProviderName returns the name of the current STT provider
func (s *STTService) GetProviderName() string {
	return s.config.Provider
}

// GetConfig returns the STT configuration.
func (s *STTService) GetConfig() *config.STTConfig {
	return s.config
}
