package repository

import (
	"database/sql"
	"time"

	"github.com/mino/backend/internal/model"
)

type SpeakerRepository struct {
	db *sql.DB
}

func NewSpeakerRepository(db *sql.DB) *SpeakerRepository {
	return &SpeakerRepository{db: db}
}

func (r *SpeakerRepository) Create(sp *model.SpeakerProfile) error {
	query := `INSERT INTO mino_speaker_profiles (user_id, name, milvus_speaker_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, sp.UserID, sp.Name, sp.MilvusSpeakerID).
		Scan(&sp.ID, &sp.CreatedAt, &sp.UpdatedAt)
}

func (r *SpeakerRepository) FindByID(id, userID string) (*model.SpeakerProfile, error) {
	sp := &model.SpeakerProfile{}
	query := `SELECT id, user_id, name, milvus_speaker_id, created_at, updated_at
		FROM mino_speaker_profiles WHERE id = $1 AND user_id = $2`
	err := r.db.QueryRow(query, id, userID).Scan(
		&sp.ID, &sp.UserID, &sp.Name, &sp.MilvusSpeakerID, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return sp, err
}

func (r *SpeakerRepository) FindByMilvusID(milvusSpeakerID string) (*model.SpeakerProfile, error) {
	sp := &model.SpeakerProfile{}
	query := `SELECT id, user_id, name, milvus_speaker_id, created_at, updated_at
		FROM mino_speaker_profiles WHERE milvus_speaker_id = $1`
	err := r.db.QueryRow(query, milvusSpeakerID).Scan(
		&sp.ID, &sp.UserID, &sp.Name, &sp.MilvusSpeakerID, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return sp, err
}

func (r *SpeakerRepository) List(userID string) ([]*model.SpeakerProfile, error) {
	query := `SELECT id, user_id, name, milvus_speaker_id, created_at, updated_at
		FROM mino_speaker_profiles WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var speakers []*model.SpeakerProfile
	for rows.Next() {
		sp := &model.SpeakerProfile{}
		if err := rows.Scan(&sp.ID, &sp.UserID, &sp.Name, &sp.MilvusSpeakerID, &sp.CreatedAt, &sp.UpdatedAt); err != nil {
			return nil, err
		}
		speakers = append(speakers, sp)
	}
	return speakers, rows.Err()
}

func (r *SpeakerRepository) UpdateName(id, userID, name string) error {
	query := `UPDATE mino_speaker_profiles SET name = $1, updated_at = $2 WHERE id = $3 AND user_id = $4`
	res, err := r.db.Exec(query, name, time.Now(), id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SpeakerRepository) Delete(id, userID string) (string, error) {
	// Return milvus_speaker_id so caller can clean up Milvus
	var milvusID string
	query := `DELETE FROM mino_speaker_profiles WHERE id = $1 AND user_id = $2 RETURNING milvus_speaker_id`
	err := r.db.QueryRow(query, id, userID).Scan(&milvusID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return milvusID, err
}
