package service

// RealtimeSession is a unified interface for real-time STT sessions
type RealtimeSession interface {
	// SendAudio sends a chunk of PCM audio (16-bit, 16kHz, mono)
	SendAudio(pcm []byte) error
	
	// Close stops the session and releases resources
	Close()
}

// NewRealtimeSession creates a real-time STT session based on the provider
func (s *STTService) NewRealtimeSession(onResult func(text string, isFinal bool)) (RealtimeSession, error) {
	switch s.config.Provider {
	case "whisper":
		return NewWhisperRealtimeSession(
			s.config.WhisperAPIURL,
			s.config.WhisperAPIKey,
			onResult,
		)
	default:
		// Return a no-op session if STT is not configured
		return &noopRealtimeSession{}, nil
	}
}

// noopRealtimeSession is a no-op implementation for when STT is disabled
type noopRealtimeSession struct{}

func (n *noopRealtimeSession) SendAudio(pcm []byte) error { return nil }
func (n *noopRealtimeSession) Close()                     {}
