package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mino/backend/internal/config"
	"github.com/mino/backend/internal/model"
	"github.com/mino/backend/internal/pkg/vectordb"
	"github.com/mino/backend/internal/repository"
	"github.com/sirupsen/logrus"
)

// SpeakerService manages speaker voice profiles — matching embeddings via Milvus
// and persisting metadata in PostgreSQL.
type SpeakerService struct {
	repo      *repository.SpeakerRepository
	milvus    *vectordb.Client
	cfg       *config.STTConfig
	logger    *logrus.Logger
}

func NewSpeakerService(
	repo *repository.SpeakerRepository,
	milvus *vectordb.Client,
	cfg *config.STTConfig,
	logger *logrus.Logger,
) *SpeakerService {
	return &SpeakerService{
		repo:   repo,
		milvus: milvus,
		cfg:    cfg,
		logger: logger,
	}
}

// SpeakerMatch holds the result of identifying a speaker from an embedding.
type SpeakerMatch struct {
	SpeakerID   string  `json:"speaker_id"`
	SpeakerName string  `json:"speaker_name"`
	Score       float32 `json:"score"`
	IsNew       bool    `json:"is_new"` // true if no known speaker matched
}

// IdentifySpeaker searches Milvus for the closest known speaker embedding.
// If the best match exceeds the similarity threshold, returns the known speaker.
// Otherwise returns IsNew=true with a placeholder name.
func (s *SpeakerService) IdentifySpeaker(ctx context.Context, userID string, embedding []float32) (*SpeakerMatch, error) {
	if s.milvus == nil {
		return &SpeakerMatch{IsNew: true, SpeakerName: "未知说话人"}, nil
	}

	hits, err := s.milvus.SearchSpeaker(ctx, userID, embedding, 1)
	if err != nil {
		s.logger.Warnf("speaker search failed: %v", err)
		return &SpeakerMatch{IsNew: true, SpeakerName: "未知说话人"}, nil
	}

	threshold := float32(s.cfg.SpeakerSimilarityThreshold)
	if threshold <= 0 {
		threshold = 0.75
	}

	if len(hits) > 0 && hits[0].Score >= threshold {
		return &SpeakerMatch{
			SpeakerID:   hits[0].SpeakerID,
			SpeakerName: hits[0].SpeakerName,
			Score:       hits[0].Score,
			IsNew:       false,
		}, nil
	}

	return &SpeakerMatch{IsNew: true, SpeakerName: "未知说话人"}, nil
}

// RegisterSpeaker creates a new speaker profile: stores embedding in Milvus and metadata in PostgreSQL.
func (s *SpeakerService) RegisterSpeaker(ctx context.Context, userID, name string, embedding []float32) (*model.SpeakerProfile, error) {
	speakerID := uuid.New().String()

	// Store embedding in Milvus
	if s.milvus != nil {
		if err := s.milvus.InsertSpeaker(ctx, userID, speakerID, name, embedding); err != nil {
			return nil, fmt.Errorf("failed to store speaker embedding: %w", err)
		}
	}

	// Store metadata in PostgreSQL
	sp := &model.SpeakerProfile{
		UserID:          userID,
		Name:            name,
		MilvusSpeakerID: speakerID,
	}
	if err := s.repo.Create(sp); err != nil {
		// Rollback Milvus insert on DB failure
		if s.milvus != nil {
			_ = s.milvus.DeleteSpeaker(ctx, speakerID)
		}
		return nil, fmt.Errorf("failed to create speaker profile: %w", err)
	}

	return sp, nil
}

// UpdateSpeakerName updates the display name of a speaker profile.
func (s *SpeakerService) UpdateSpeakerName(ctx context.Context, userID, speakerProfileID, newName string) error {
	return s.repo.UpdateName(speakerProfileID, userID, newName)
}

// ListSpeakers returns all speaker profiles for a user.
func (s *SpeakerService) ListSpeakers(ctx context.Context, userID string) ([]*model.SpeakerProfile, error) {
	return s.repo.List(userID)
}

// DeleteSpeaker removes a speaker profile from both PostgreSQL and Milvus.
func (s *SpeakerService) DeleteSpeaker(ctx context.Context, userID, speakerProfileID string) error {
	milvusID, err := s.repo.Delete(speakerProfileID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete speaker profile: %w", err)
	}
	if milvusID != "" && s.milvus != nil {
		if err := s.milvus.DeleteSpeaker(ctx, milvusID); err != nil {
			s.logger.Warnf("failed to delete speaker from Milvus (id=%s): %v", milvusID, err)
		}
	}
	return nil
}

// ResolveSpeakers takes diarized segments with raw speaker labels (SPEAKER_00, etc.)
// and their embeddings, identifies each speaker against known profiles, and returns
// segments with resolved speaker names.
func (s *SpeakerService) ResolveSpeakers(
	ctx context.Context,
	userID string,
	segments []model.DiarizedSegment,
	speakerEmbeddings map[string][]float32,
) ([]model.DiarizedSegment, map[string]*SpeakerMatch) {
	speakerMap := make(map[string]*SpeakerMatch)
	unknownCounter := 0

	for label, embedding := range speakerEmbeddings {
		match, err := s.IdentifySpeaker(ctx, userID, embedding)
		if err != nil {
			s.logger.Warnf("failed to identify speaker %s: %v", label, err)
			unknownCounter++
			match = &SpeakerMatch{
				IsNew:       true,
				SpeakerName: fmt.Sprintf("未知说话人_%d", unknownCounter),
			}
		}
		if match.IsNew && match.SpeakerName == "未知说话人" {
			unknownCounter++
			match.SpeakerName = fmt.Sprintf("未知说话人_%d", unknownCounter)
		}
		speakerMap[label] = match
	}

	// Apply resolved names to segments
	resolved := make([]model.DiarizedSegment, len(segments))
	for i, seg := range segments {
		resolved[i] = seg
		if m, ok := speakerMap[seg.Speaker]; ok {
			resolved[i].SpeakerName = m.SpeakerName
		} else {
			resolved[i].SpeakerName = seg.Speaker
		}
	}

	return resolved, speakerMap
}
