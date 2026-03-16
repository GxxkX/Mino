package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	"github.com/mino/backend/internal/model"
)

// WhisperSTT implements STTProvider for local Whisper API
type WhisperSTT struct {
	baseURL    string
	apiKey     string
	model      string
	language   string // e.g. "zh", "en", "" for auto-detect
	httpClient *http.Client
}

// NewWhisperSTT creates a new Whisper STT provider
func NewWhisperSTT(apiURL, apiKey, model, language string) *WhisperSTT {
	if apiURL == "" {
		apiURL = "http://localhost:9000"
	}
	if model == "" {
		model = "turbo"
	}
	
	return &WhisperSTT{
		baseURL:    apiURL,
		apiKey:     apiKey,
		model:      model,
		language:   language,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// TranscribeStream performs real-time streaming transcription using /transcribe_stream endpoint
func (w *WhisperSTT) TranscribeStream(ctx context.Context, audioStream io.Reader) (<-chan TranscriptResult, error) {
	resultChan := make(chan TranscriptResult, 10)
	
	go func() {
		defer close(resultChan)
		
		// Read all audio data
		audioData, err := io.ReadAll(audioStream)
		if err != nil {
			resultChan <- TranscriptResult{Error: fmt.Errorf("failed to read audio stream: %w", err)}
			return
		}
		
		// Use stream transcription endpoint
		text, err := w.transcribeWithEndpoint(ctx, audioData, "webm", "/transcribe_stream")
		if err != nil {
			resultChan <- TranscriptResult{Error: err}
			return
		}
		
		// Send final result
		resultChan <- TranscriptResult{
			Text:       text,
			IsFinal:    true,
			Timestamp:  time.Now().UnixMilli(),
			Confidence: 1.0,
		}
	}()
	
	return resultChan, nil
}

// TranscribeFile performs batch transcription on a complete audio file using /transcribe endpoint
func (w *WhisperSTT) TranscribeFile(ctx context.Context, audioData []byte, format string) (string, error) {
	return w.transcribeWithEndpoint(ctx, audioData, format, "/transcribe")
}

// transcribeWithEndpoint performs transcription using the specified endpoint
func (w *WhisperSTT) transcribeWithEndpoint(ctx context.Context, audioData []byte, format string, endpoint string) (string, error) {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Add file field with explicit Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="audio.%s"`, format))
	h.Set("Content-Type", "audio/"+format)
	part, err := writer.CreatePart(h)
	if err != nil {
		return "", fmt.Errorf("failed to create form part: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	// Add language field if specified
	if w.language != "" {
		if err := writer.WriteField("language", w.language); err != nil {
			return "", fmt.Errorf("failed to write language field: %w", err)
		}
	}
	
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}
	
	// Create HTTP request
	apiURL := w.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}
	
	// Send request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	// Parse response from whisper_api.py format
	var result struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Segments []struct {
			Text  string  `json:"text"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		} `json:"segments"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Text, nil
}

// DiarizeFile performs speaker diarization + transcription via the /diarize endpoint.
func (w *WhisperSTT) DiarizeFile(ctx context.Context, audioData []byte, format string) (*DiarizedResult, error) {
	// Use a longer timeout for diarization (it's heavier than plain transcription)
	client := &http.Client{Timeout: 180 * time.Second}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="audio.%s"`, format))
	h.Set("Content-Type", "audio/"+format)
	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("failed to create form part: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return nil, fmt.Errorf("failed to write audio data: %w", err)
	}

	// Add language field if specified
	if w.language != "" {
		if err := writer.WriteField("language", w.language); err != nil {
			return nil, fmt.Errorf("failed to write language field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	apiURL := w.baseURL + "/diarize"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send diarize request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("diarize API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResult struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Segments []struct {
			Speaker string  `json:"speaker"`
			Text    string  `json:"text"`
			Start   float64 `json:"start"`
			End     float64 `json:"end"`
		} `json:"segments"`
		Speakers    map[string]struct {
			Embedding []float32 `json:"embedding"`
		} `json:"speakers"`
		NumSpeakers int `json:"num_speakers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
		return nil, fmt.Errorf("failed to decode diarize response: %w", err)
	}

	// Convert to internal types
	segments := make([]model.DiarizedSegment, len(apiResult.Segments))
	for i, seg := range apiResult.Segments {
		segments[i] = model.DiarizedSegment{
			Speaker: seg.Speaker,
			Text:    seg.Text,
			Start:   seg.Start,
			End:     seg.End,
		}
	}

	embeddings := make(map[string][]float32)
	for label, spk := range apiResult.Speakers {
		if len(spk.Embedding) > 0 {
			embeddings[label] = spk.Embedding
		}
	}

	return &DiarizedResult{
		Text:              apiResult.Text,
		Language:          apiResult.Language,
		Segments:          segments,
		SpeakerEmbeddings: embeddings,
		NumSpeakers:       apiResult.NumSpeakers,
	}, nil
}

// Close releases any resources held by the provider
func (w *WhisperSTT) Close() error {
	// No resources to clean up for HTTP-based client
	return nil
}
