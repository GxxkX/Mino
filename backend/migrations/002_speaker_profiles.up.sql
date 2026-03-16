-- 002_speaker_profiles.up.sql

CREATE TABLE IF NOT EXISTS mino_speaker_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    milvus_speaker_id VARCHAR(64) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_speaker_profiles_user_id ON mino_speaker_profiles(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_speaker_profiles_milvus_id ON mino_speaker_profiles(milvus_speaker_id);

-- Add diarized_transcript column to conversations for storing speaker-segmented transcript
ALTER TABLE mino_conversations ADD COLUMN IF NOT EXISTS diarized_transcript JSONB;
