-- 002_speaker_profiles.down.sql

ALTER TABLE mino_conversations DROP COLUMN IF EXISTS diarized_transcript;
DROP TABLE IF EXISTS mino_speaker_profiles;
