package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

// WhisperRealtimeSession accumulates audio chunks and periodically sends them
// to the Whisper API for transcription. This simulates real-time transcription
// by batching audio data.
type WhisperRealtimeSession struct {
	baseURL    string
	apiKey     string
	onResult   func(text string, isFinal bool)
	
	mu         sync.Mutex
	audioData  []byte
	closed     bool
	httpClient *http.Client
	
	// Batch processing
	batchSize     int           // bytes to accumulate before sending
	batchInterval time.Duration // max time to wait before sending
	lastSent      time.Time
	timer         *time.Timer
	
	// Text accumulation for incremental results
	accumulatedText string
}

// NewWhisperRealtimeSession creates a new Whisper real-time session
func NewWhisperRealtimeSession(baseURL, apiKey string, onResult func(text string, isFinal bool)) (*WhisperRealtimeSession, error) {
	if baseURL == "" {
		baseURL = "http://localhost:9000"
	}
	
	s := &WhisperRealtimeSession{
		baseURL:       baseURL,
		apiKey:        apiKey,
		onResult:      onResult,
		audioData:     make([]byte, 0, 1024*1024), // 1MB initial capacity
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		batchSize:     160000, // ~5 seconds of 16kHz 16-bit mono PCM (16000 * 2 * 5)
		batchInterval: 3 * time.Second,
		lastSent:      time.Now(),
	}
	
	// Start periodic batch processing
	s.timer = time.AfterFunc(s.batchInterval, s.processBatch)
	
	return s, nil
}

// SendAudio accumulates audio data and triggers transcription when batch size is reached
func (s *WhisperRealtimeSession) SendAudio(pcm []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.closed {
		return fmt.Errorf("session is closed")
	}
	
	s.audioData = append(s.audioData, pcm...)
	
	// If we've accumulated enough data, process immediately
	if len(s.audioData) >= s.batchSize {
		go s.processBatch()
	}
	
	return nil
}

// processBatch sends accumulated audio to Whisper API
func (s *WhisperRealtimeSession) processBatch() {
	s.mu.Lock()
	if s.closed || len(s.audioData) == 0 {
		s.mu.Unlock()
		return
	}
	
	// Take current audio data
	audioToSend := make([]byte, len(s.audioData))
	copy(audioToSend, s.audioData)
	s.audioData = s.audioData[:0] // Clear buffer
	s.lastSent = time.Now()
	s.mu.Unlock()
	
	// Reset timer for next batch
	if s.timer != nil {
		s.timer.Reset(s.batchInterval)
	}
	
	// Send to Whisper API (non-blocking)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	
	text, err := s.transcribe(ctx, audioToSend)
	if err != nil {
		log.Printf("whisper realtime: transcription error: %v", err)
		return
	}
	
	if text != "" {
		s.mu.Lock()
		// 计算增量文本（新转录的部分）
		newText := text
		if len(s.accumulatedText) > 0 {
			// 如果新文本包含已累积的文本，只发送新增部分
			if len(text) > len(s.accumulatedText) {
				newText = text[len(s.accumulatedText):]
				// 去除开头的空格
				newText = strings.TrimLeft(newText, " ")
			} else {
				// 新文本更短，可能是新的句子，直接使用
				newText = text
			}
		}
		s.accumulatedText = text
		s.mu.Unlock()
		
		if newText != "" {
			// 发送增量结果
			s.onResult(newText, false)
		}
	}
}

// transcribe sends audio to Whisper /transcribe_stream endpoint
func (s *WhisperRealtimeSession) transcribe(ctx context.Context, audioData []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Add file field with explicit Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="audio.wav"`)
	h.Set("Content-Type", "audio/wav")
	part, err := writer.CreatePart(h)
	if err != nil {
		return "", fmt.Errorf("failed to create form part: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}
	
	// Create HTTP request to /transcribe_stream endpoint
	apiURL := s.baseURL + "/transcribe_stream"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	
	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var result struct {
		Text     string `json:"text"`
		Language string `json:"language"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Text, nil
}

// Close stops the session and processes any remaining audio
func (s *WhisperRealtimeSession) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	
	// Stop timer
	if s.timer != nil {
		s.timer.Stop()
	}
	
	// Process any remaining audio
	remainingAudio := make([]byte, len(s.audioData))
	copy(remainingAudio, s.audioData)
	finalAccumulatedText := s.accumulatedText
	s.audioData = nil
	s.mu.Unlock()
	
	// Send final batch if there's any audio left
	if len(remainingAudio) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		text, err := s.transcribe(ctx, remainingAudio)
		if err != nil {
			log.Printf("whisper realtime: final transcription error: %v", err)
			return
		}
		
		if text != "" {
			// 计算最后的增量文本
			newText := text
			if len(finalAccumulatedText) > 0 && len(text) > len(finalAccumulatedText) {
				newText = text[len(finalAccumulatedText):]
				newText = strings.TrimLeft(newText, " ")
			}
			
			if newText != "" {
				// Send as final result
				s.onResult(newText, true)
			}
		}
	}
}
