-- 001_init.up.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS mino_users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    email           VARCHAR(255) UNIQUE,
    display_name    VARCHAR(100),
    avatar_url      TEXT,
    role            VARCHAR(20) DEFAULT 'user',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mino_conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    title           VARCHAR(255),
    summary         TEXT,
    transcript      TEXT NOT NULL DEFAULT '',
    audio_url       TEXT,
    audio_duration  INTEGER,
    language        VARCHAR(10) DEFAULT 'zh',
    status          VARCHAR(20) DEFAULT 'completed',
    recorded_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON mino_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_recorded_at ON mino_conversations(recorded_at);

CREATE TABLE IF NOT EXISTS mino_memories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES mino_conversations(id) ON DELETE SET NULL,
    content         TEXT NOT NULL,
    category        VARCHAR(50),
    importance      INTEGER DEFAULT 5,
    embedding_id    BIGINT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_memories_user_id ON mino_memories(user_id);
CREATE INDEX IF NOT EXISTS idx_memories_category ON mino_memories(category);

CREATE TABLE IF NOT EXISTS mino_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES mino_conversations(id) ON DELETE SET NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    status          VARCHAR(20) DEFAULT 'pending',
    priority        VARCHAR(10) DEFAULT 'medium',
    due_date        TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON mino_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON mino_tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON mino_tasks(due_date);

CREATE TABLE IF NOT EXISTS mino_tags (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,
    color           VARCHAR(7),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON mino_tags(user_id);

CREATE TABLE IF NOT EXISTS mino_conversation_tags (
    conversation_id UUID REFERENCES mino_conversations(id) ON DELETE CASCADE,
    tag_id          UUID REFERENCES mino_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (conversation_id, tag_id)
);

CREATE TABLE IF NOT EXISTS mino_chat_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL DEFAULT '新对话',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_id ON mino_chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON mino_chat_sessions(updated_at);

CREATE TABLE IF NOT EXISTS mino_chat_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES mino_chat_sessions(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL,
    content         TEXT NOT NULL,
    sources         JSONB,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_session_id ON mino_chat_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON mino_chat_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON mino_chat_messages(created_at);

CREATE TABLE IF NOT EXISTS mino_extensions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES mino_users(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    icon            VARCHAR(50) NOT NULL DEFAULT 'zap',
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    config          JSONB,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_extensions_user_id ON mino_extensions(user_id);
