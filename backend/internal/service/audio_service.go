package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mino/backend/internal/config"
	"github.com/mino/backend/internal/model" // used for Conversation
	"github.com/mino/backend/internal/pkg/storage"
	"github.com/mino/backend/internal/repository"
)

type AudioService struct {
	convRepo       *repository.ConversationRepository
	memTaskService *MemoryTaskService
	llmService     LLMProvider
	vectorStore    *VectorStoreService // nil if vector search is unavailable
	storage        *storage.Client
	cfg            *config.Config
	logger         *logrus.Logger
}

func NewAudioService(
	convRepo *repository.ConversationRepository,
	memTaskService *MemoryTaskService,
	llmService LLMProvider,
	vectorStore *VectorStoreService,
	storageClient *storage.Client,
	cfg *config.Config,
	logger *logrus.Logger,
) *AudioService {
	return &AudioService{
		convRepo:       convRepo,
		memTaskService: memTaskService,
		llmService:     llmService,
		vectorStore:    vectorStore,
		storage:        storageClient,
		cfg:            cfg,
		logger:         logger,
	}
}

// ProcessTranscript takes a completed transcript + raw audio, runs LLM extraction,
// uploads audio to MinIO, and persists everything.
func (s *AudioService) ProcessTranscript(ctx context.Context, userID, transcript string, audioDuration int, audioData []byte) (*model.Conversation, error) {
	now := time.Now()

	conv := &model.Conversation{
		UserID:        userID,
		Transcript:    transcript,
		Language:      "zh",
		Status:        "processing",
		RecordedAt:    &now,
		AudioDuration: &audioDuration,
	}
	if err := s.convRepo.Create(conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Upload audio to MinIO (if we have data and a storage client).
	// Audio data can be:
	// - WebM format from web MediaRecorder (opus-encoded)
	// - Raw PCM from WebSocket real-time recording (16kHz, 16-bit, mono)
	if s.storage != nil && len(audioData) > 0 {
		// Detect format based on magic bytes.
		// WebM: starts with 0x1A 0x45 0xDF 0xA3 (EBML header)
		// PCM: raw 16-bit little-endian samples — wrap in WAV so browsers can play it.
		var objectKey, contentType string
		var uploadData []byte
		if len(audioData) > 4 && audioData[0] == 0x1A && audioData[1] == 0x45 {
			// WebM/Opus — upload as-is
			objectKey = fmt.Sprintf("%s/%s.webm", userID, conv.ID)
			contentType = "audio/webm"
			uploadData = audioData
		} else {
			// Raw PCM (16kHz, 16-bit, mono) — wrap in WAV container so the
			// browser <audio> element can play it natively.
			objectKey = fmt.Sprintf("%s/%s.wav", userID, conv.ID)
			contentType = "audio/wav"
			uploadData = pcmToWAV(audioData, 16000, 1, 16)
		}

		audioURL, err := s.storage.UploadAudio(
			ctx,
			objectKey,
			bytes.NewReader(uploadData),
			int64(len(uploadData)),
			contentType,
		)
		if err != nil {
			// Log but don't fail the whole operation
			s.logger.Warnf("failed to upload audio to MinIO: %v", err)
		} else {
			conv.AudioURL = &audioURL
			s.convRepo.Update(conv)
		}
	}

	// Run LLM extraction synchronously so the caller gets the full result
	// (title, summary, etc.) before sending the completed message.
	result, err := s.llmService.ExtractAgent(ctx, transcript)
	if err != nil {
		s.logger.Warnf("LLM extraction failed: %v", err)
		conv.Status = "completed"
		s.convRepo.Update(conv)
		return conv, nil
	}

	conv.Title = &result.Title
	conv.Summary = &result.Summary
	conv.Status = "completed"
	s.convRepo.Update(conv)

	// Use MemoryTaskService for memory and task creation (shared with ChatAgent tools).
	if len(result.Memories) > 0 {
		var memInputs []MemoryInput
		for _, m := range result.Memories {
			memInputs = append(memInputs, MemoryInput{
				Content:    m.Content,
				Category:   m.Category,
				Importance: m.Importance,
			})
		}
		// CreateMemories handles DB persistence + vector store indexing internally.
		if _, err := s.memTaskService.CreateMemories(ctx, userID, memInputs, &conv.ID); err != nil {
			s.logger.Warnf("failed to create memories: %v", err)
		}
	}

	if len(result.ActionItems) > 0 {
		var taskInputs []TaskInput
		for _, item := range result.ActionItems {
			taskInputs = append(taskInputs, TaskInput{
				Title:    item,
				Priority: "medium",
			})
		}
		if _, err := s.memTaskService.CreateTasks(ctx, userID, taskInputs, &conv.ID); err != nil {
			s.logger.Warnf("failed to create tasks: %v", err)
		}
	}

	// Index conversation into Milvus for semantic search (async, non-blocking).
	// Memory indexing is already handled by MemoryTaskService.CreateMemories.
	if s.vectorStore != nil {
		go func() {
			bgCtx := context.Background()
			textToIndex := transcript
			if result.Summary != "" {
				textToIndex = result.Summary + "\n\n" + transcript
			}
			if err := s.vectorStore.IndexConversation(bgCtx, userID, conv.ID, textToIndex); err != nil {
				s.logger.Warnf("failed to index conversation in Milvus: %v", err)
			}
		}()
	}

	return conv, nil
}

// pcmToWAV wraps raw PCM samples in a standard WAV (RIFF) container.
// Parameters: sampleRate (e.g. 16000), numChannels (1=mono), bitsPerSample (16).
func pcmToWAV(pcm []byte, sampleRate, numChannels, bitsPerSample int) []byte {
	dataSize := len(pcm)
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8

	buf := new(bytes.Buffer)
	// RIFF chunk
	buf.WriteString("RIFF")
	writeUint32LE(buf, uint32(36+dataSize))
	buf.WriteString("WAVE")
	// fmt sub-chunk
	buf.WriteString("fmt ")
	writeUint32LE(buf, 16) // PCM chunk size
	writeUint16LE(buf, 1)  // AudioFormat = PCM
	writeUint16LE(buf, uint16(numChannels))
	writeUint32LE(buf, uint32(sampleRate))
	writeUint32LE(buf, uint32(byteRate))
	writeUint16LE(buf, uint16(blockAlign))
	writeUint16LE(buf, uint16(bitsPerSample))
	// data sub-chunk
	buf.WriteString("data")
	writeUint32LE(buf, uint32(dataSize))
	buf.Write(pcm)
	return buf.Bytes()
}

func writeUint32LE(b *bytes.Buffer, v uint32) {
	b.WriteByte(byte(v))
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v >> 16))
	b.WriteByte(byte(v >> 24))
}

func writeUint16LE(b *bytes.Buffer, v uint16) {
	b.WriteByte(byte(v))
	b.WriteByte(byte(v >> 8))
}

// UploadAudio stores an audio file to MinIO and returns the URL.
func (s *AudioService) UploadAudio(ctx context.Context, userID, convID string, r io.Reader, size int64, contentType string) (string, error) {
	if s.storage == nil {
		return fmt.Sprintf("/audio/%s/%s", userID, convID), nil
	}
	objectKey := fmt.Sprintf("%s/%s.webm", userID, convID)
	return s.storage.UploadAudio(ctx, objectKey, r, size, contentType)
}
