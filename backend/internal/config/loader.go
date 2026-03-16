package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Load reads configuration from environment variables (and optionally a .env file)
func Load(envFile string) (*Config, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			// Not fatal — env vars may already be set
			fmt.Printf("Warning: could not load %s: %v\n", envFile, err)
		}
	}

	accessExpire, err := time.ParseDuration(getEnv("JWT_ACCESS_TOKEN_EXPIRE", "15m"))
	if err != nil {
		accessExpire = 15 * time.Minute
	}
	refreshExpire, err := time.ParseDuration(getEnv("JWT_REFRESH_TOKEN_EXPIRE", "168h"))
	if err != nil {
		refreshExpire = 168 * time.Hour
	}

	cfg := &Config{
		App: AppConfig{
			Env:   getEnv("APP_ENV", "development"),
			Port:  getEnv("APP_PORT", "8000"),
			Debug: getEnvBool("APP_DEBUG", true),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "mino"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			PrivateKeyPath:     getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
			PublicKeyPath:      getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
			AccessTokenExpire:  accessExpire,
			RefreshTokenExpire: refreshExpire,
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "mino"),
			Password: getEnv("ADMIN_PASSWORD", "admin"),
		},
		LLM: LLMConfig{
			Provider:       getEnv("LLM_PROVIDER", "openai"),
			APIKey:         getEnv("LLM_API_KEY", ""),
			BaseURL:        getEnv("LLM_BASE_URL", ""),
			Model:          getEnv("LLM_MODEL", "gpt-4o"),
			EmbeddingModel: getEnv("LLM_EMBEDDING_MODEL", "embedding-3"),
		},
		Milvus: MilvusConfig{
			Host:                       getEnv("MILVUS_HOST", "localhost"),
			Port:                       getEnv("MILVUS_PORT", "19530"),
			User:                       getEnv("MILVUS_USER", ""),
			Password:                   getEnv("MILVUS_PASSWORD", ""),
			DBName:                     getEnv("MILVUS_DB_NAME", "default"),
			ConversationsCollection:    getEnv("MILVUS_CONVERSATIONS_COLLECTION", "conversations"),
			MemoriesCollection:         getEnv("MILVUS_MEMORIES_COLLECTION", "memories"),
			SpeakerEmbeddingsCollection: getEnv("MILVUS_SPEAKER_EMBEDDINGS_COLLECTION", "speaker_embeddings"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", ""),
			SecretKey: getEnv("MINIO_SECRET_KEY", ""),
			Secure:    getEnvBool("MINIO_SECURE", false),
			Region:    getEnv("MINIO_REGION", "us-east-1"),
			PublicURL: getEnv("MINIO_PUBLIC_URL", ""),
		},
		Typesense: TypesenseConfig{
			Host:   getEnv("TYPESENSE_HOST", "localhost"),
			Port:   getEnv("TYPESENSE_PORT", "8108"),
			APIKey: getEnv("TYPESENSE_API_KEY", ""),
		},
		LangSmith: LangSmithConfig{
			Tracing:               getEnvBool("LANGSMITH_TRACING", false),
			APIKey:                getEnv("LANGSMITH_API_KEY", ""),
			Project:               getEnv("LANGSMITH_PROJECT", "mino-backend-chat"),
			Endpoint:              getEnv("LANGSMITH_ENDPOINT", "https://api.smith.langchain.com"),
			AgenticPromptName:     getEnv("OMI_LANGSMITH_AGENTIC_PROMPT_NAME", "mino-agentic-system"),
			PromptCacheTTLSeconds: getEnvInt("OMI_LANGSMITH_PROMPT_CACHE_TTL_SECONDS", 300),
		},
		STT: STTConfig{
			Provider:                   getEnv("STT_PROVIDER", "whisper"),
			WhisperAPIURL:              getEnv("STT_WHISPER_API_URL", "http://localhost:9000"),
			WhisperAPIKey:              getEnv("STT_WHISPER_API_KEY", ""),
			WhisperModel:               getEnv("STT_WHISPER_MODEL", "turbo"),
			WhisperLanguage:            getEnv("STT_WHISPER_LANGUAGE", ""),
			PyannoteEnabled:            getEnvBool("STT_PYANNOTE_ENABLED", false),
			SpeakerSimilarityThreshold: getEnvFloat("SPEAKER_SIMILARITY_THRESHOLD", 0.75),
		},
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}
