package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/config"
	"github.com/mino/backend/internal/pkg/response"
)

// SettingsHandler exposes runtime-configurable settings via the API.
type SettingsHandler struct {
	cfg *config.Config
}

func NewSettingsHandler(cfg *config.Config) *SettingsHandler {
	return &SettingsHandler{cfg: cfg}
}

// LLMConfigResponse is the JSON shape returned / accepted for LLM settings.
type LLMConfigResponse struct {
	Provider       string `json:"provider"`
	APIKey         string `json:"api_key"`
	BaseURL        string `json:"base_url"`
	Model          string `json:"model"`
	EmbeddingModel string `json:"embedding_model"`
}

// GetLLMConfig returns the current LLM configuration.
// Sensitive fields (api_key) are returned empty — the client shows them as blank.
// GET /v1/settings/llm
func (h *SettingsHandler) GetLLMConfig(c *gin.Context) {
	resp := LLMConfigResponse{
		Provider:       h.cfg.LLM.Provider,
		APIKey:         "", // never expose
		BaseURL:        h.cfg.LLM.BaseURL,
		Model:          h.cfg.LLM.Model,
		EmbeddingModel: h.cfg.LLM.EmbeddingModel,
	}
	response.OK(c, resp)
}

// UpdateLLMConfig updates the LLM configuration at runtime.
// PUT /v1/settings/llm
func (h *SettingsHandler) UpdateLLMConfig(c *gin.Context) {
	var req LLMConfigResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if req.Provider != "" {
		h.cfg.LLM.Provider = req.Provider
	}
	// Only update API key if a non-empty value is provided (empty = keep current)
	if req.APIKey != "" {
		h.cfg.LLM.APIKey = req.APIKey
	}
	if req.BaseURL != "" {
		h.cfg.LLM.BaseURL = req.BaseURL
	}
	if req.Model != "" {
		h.cfg.LLM.Model = req.Model
	}
	if req.EmbeddingModel != "" {
		h.cfg.LLM.EmbeddingModel = req.EmbeddingModel
	}

	resp := LLMConfigResponse{
		Provider:       h.cfg.LLM.Provider,
		APIKey:         "", // never expose
		BaseURL:        h.cfg.LLM.BaseURL,
		Model:          h.cfg.LLM.Model,
		EmbeddingModel: h.cfg.LLM.EmbeddingModel,
	}
	response.OK(c, resp)
}

// CloudConfigResponse is the JSON shape for cloud/infrastructure settings.
type CloudConfigResponse struct {
	// MinIO
	MinIOEndpoint  string `json:"minio_endpoint"`
	MinIOAccessKey string `json:"minio_access_key"`
	MinIOSecretKey string `json:"minio_secret_key"`
	MinIOSecure    bool   `json:"minio_secure"`
	MinIORegion    string `json:"minio_region"`
	MinIOPublicURL string `json:"minio_public_url"`

	// PostgreSQL
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBSSLMode  string `json:"db_ssl_mode"`

	// Redis
	RedisHost     string `json:"redis_host"`
	RedisPort     string `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// Milvus
	MilvusHost     string `json:"milvus_host"`
	MilvusPort     string `json:"milvus_port"`
	MilvusUser     string `json:"milvus_user"`
	MilvusPassword string `json:"milvus_password"`
	MilvusDBName   string `json:"milvus_db_name"`

	// Typesense
	TypesenseHost   string `json:"typesense_host"`
	TypesensePort   string `json:"typesense_port"`
	TypesenseAPIKey string `json:"typesense_api_key"`
}

// GetCloudConfig returns the current cloud/infrastructure configuration.
// Sensitive fields (passwords, keys) are returned empty — the client shows them as blank.
// GET /v1/settings/cloud
func (h *SettingsHandler) GetCloudConfig(c *gin.Context) {
	resp := CloudConfigResponse{
		MinIOEndpoint:  h.cfg.MinIO.Endpoint,
		MinIOAccessKey: "", // never expose
		MinIOSecretKey: "", // never expose
		MinIOSecure:    h.cfg.MinIO.Secure,
		MinIORegion:    h.cfg.MinIO.Region,
		MinIOPublicURL: h.cfg.MinIO.PublicURL,

		DBHost:     h.cfg.DB.Host,
		DBPort:     h.cfg.DB.Port,
		DBName:     h.cfg.DB.Name,
		DBUser:     h.cfg.DB.User,
		DBPassword: "", // never expose
		DBSSLMode:  h.cfg.DB.SSLMode,

		RedisHost:     h.cfg.Redis.Host,
		RedisPort:     h.cfg.Redis.Port,
		RedisPassword: "", // never expose
		RedisDB:       h.cfg.Redis.DB,

		MilvusHost:     h.cfg.Milvus.Host,
		MilvusPort:     h.cfg.Milvus.Port,
		MilvusUser:     h.cfg.Milvus.User,
		MilvusPassword: "", // never expose
		MilvusDBName:   h.cfg.Milvus.DBName,

		TypesenseHost:   h.cfg.Typesense.Host,
		TypesensePort:   h.cfg.Typesense.Port,
		TypesenseAPIKey: "", // never expose
	}
	response.OK(c, resp)
}

// UpdateCloudConfig updates the cloud/infrastructure configuration at runtime.
// PUT /v1/settings/cloud
func (h *SettingsHandler) UpdateCloudConfig(c *gin.Context) {
	var req CloudConfigResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// MinIO
	if req.MinIOEndpoint != "" {
		h.cfg.MinIO.Endpoint = req.MinIOEndpoint
	}
	if req.MinIOAccessKey != "" {
		h.cfg.MinIO.AccessKey = req.MinIOAccessKey
	}
	if req.MinIOSecretKey != "" {
		h.cfg.MinIO.SecretKey = req.MinIOSecretKey
	}
	h.cfg.MinIO.Secure = req.MinIOSecure
	if req.MinIORegion != "" {
		h.cfg.MinIO.Region = req.MinIORegion
	}
	if req.MinIOPublicURL != "" {
		h.cfg.MinIO.PublicURL = req.MinIOPublicURL
	}

	// DB
	if req.DBHost != "" {
		h.cfg.DB.Host = req.DBHost
	}
	if req.DBPort != "" {
		h.cfg.DB.Port = req.DBPort
	}
	if req.DBName != "" {
		h.cfg.DB.Name = req.DBName
	}
	if req.DBUser != "" {
		h.cfg.DB.User = req.DBUser
	}
	if req.DBPassword != "" {
		h.cfg.DB.Password = req.DBPassword
	}
	if req.DBSSLMode != "" {
		h.cfg.DB.SSLMode = req.DBSSLMode
	}

	// Redis
	if req.RedisHost != "" {
		h.cfg.Redis.Host = req.RedisHost
	}
	if req.RedisPort != "" {
		h.cfg.Redis.Port = req.RedisPort
	}
	if req.RedisPassword != "" {
		h.cfg.Redis.Password = req.RedisPassword
	}
	h.cfg.Redis.DB = req.RedisDB

	// Milvus
	if req.MilvusHost != "" {
		h.cfg.Milvus.Host = req.MilvusHost
	}
	if req.MilvusPort != "" {
		h.cfg.Milvus.Port = req.MilvusPort
	}
	if req.MilvusUser != "" {
		h.cfg.Milvus.User = req.MilvusUser
	}
	if req.MilvusPassword != "" {
		h.cfg.Milvus.Password = req.MilvusPassword
	}
	if req.MilvusDBName != "" {
		h.cfg.Milvus.DBName = req.MilvusDBName
	}

	// Typesense
	if req.TypesenseHost != "" {
		h.cfg.Typesense.Host = req.TypesenseHost
	}
	if req.TypesensePort != "" {
		h.cfg.Typesense.Port = req.TypesensePort
	}
	if req.TypesenseAPIKey != "" {
		h.cfg.Typesense.APIKey = req.TypesenseAPIKey
	}

	// Return updated config
	h.GetCloudConfig(c)
}
