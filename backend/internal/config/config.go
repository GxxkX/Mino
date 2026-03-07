package config

import "time"

// Config holds all application configuration
type Config struct {
	App       AppConfig
	DB        DBConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Admin     AdminConfig
	LLM       LLMConfig
	Milvus    MilvusConfig
	MinIO     MinIOConfig
	Typesense TypesenseConfig
	LangSmith LangSmithConfig
	STT       STTConfig
}

type AppConfig struct {
	Env   string
	Port  string
	Debug bool
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	PrivateKeyPath     string
	PublicKeyPath      string
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration
}

type AdminConfig struct {
	Username string
	Password string
}

type LLMConfig struct {
	Provider       string
	APIKey         string
	BaseURL        string
	Model          string
	EmbeddingModel string
}

type MilvusConfig struct {
	Host                    string
	Port                    string
	User                    string
	Password                string
	DBName                  string
	ConversationsCollection string
	MemoriesCollection      string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
	Region    string
	PublicURL string
}

type TypesenseConfig struct {
	Host   string
	Port   string
	APIKey string
}

type LangSmithConfig struct {
	Tracing               bool
	APIKey                string
	Project               string
	Endpoint              string
	AgenticPromptName     string
	PromptCacheTTLSeconds int
}

type STTConfig struct {
	Provider string // "whisper"

	// Whisper configuration
	WhisperAPIURL string
	WhisperAPIKey string
	WhisperModel  string
}
