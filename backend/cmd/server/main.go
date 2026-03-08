package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/mino/backend/internal/config"
	"github.com/mino/backend/migrations"
	"github.com/mino/backend/internal/handler"
	"github.com/mino/backend/internal/middleware"
	jwtpkg "github.com/mino/backend/internal/pkg/jwt"
	"github.com/mino/backend/internal/pkg/search"
	"github.com/mino/backend/internal/pkg/storage"
	"github.com/mino/backend/internal/pkg/vectordb"
	"github.com/mino/backend/internal/repository"
	"github.com/mino/backend/internal/service"
	"github.com/mino/backend/internal/service/tools"
)

func main() {
	// Load config
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	cfg, err := config.Load(envFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	if cfg.App.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Database
	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	logger.Info("connected to PostgreSQL")

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// JWT Manager
	jwtMgr, err := jwtpkg.NewManager(
		cfg.JWT.PrivateKeyPath,
		cfg.JWT.PublicKeyPath,
		cfg.JWT.AccessTokenExpire,
		cfg.JWT.RefreshTokenExpire,
	)
	if err != nil {
		log.Fatalf("failed to initialize JWT manager: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	memRepo := repository.NewMemoryRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	chatRepo := repository.NewChatRepository(db)
	extRepo := repository.NewExtensionRepository(db)

	// Services
	// Use LangchainLLMService for LLM integration with LangSmith tracing
	llmProvider, err := service.NewLangchainLLMService(&cfg.LLM, &cfg.LangSmith, logger)
	if err != nil {
		logger.Fatalf("failed to initialize LangchainLLMService: %v", err)
	}
	logger.Info("using LangchainLLMService with LangSmith tracing")

	// Milvus vector database client (optional — app works without it)
	var milvusClient *vectordb.Client
	if cfg.Milvus.Host != "" {
		mc, err := vectordb.NewClient(&cfg.Milvus, logger)
		if err != nil {
			logger.Warnf("Milvus init failed (vector search disabled): %v", err)
		} else {
			milvusClient = mc
			defer milvusClient.Close()
			// Ensure collections exist and are loaded
			if err := milvusClient.EnsureCollections(context.Background()); err != nil {
				logger.Warnf("Milvus collection init failed: %v", err)
			} else {
				logger.Info("Milvus collections ready")
			}
		}
	}

	// VectorStore service (nil if Milvus or embedder unavailable)
	vectorStoreSvc := service.NewVectorStoreService(milvusClient, llmProvider, logger)

	authSvc := service.NewAuthService(userRepo, jwtMgr, cfg)

	// MinIO storage client (optional — recording works without it)
	var storageClient *storage.Client
	if cfg.MinIO.Endpoint != "" && cfg.MinIO.AccessKey != "" {
		sc, err := storage.NewClient(&cfg.MinIO)
		if err != nil {
			logger.Warnf("MinIO init failed (audio upload disabled): %v", err)
		} else {
			storageClient = sc
			logger.Info("connected to MinIO")
		}
	}

	// STT service (optional — recording works without it, but transcription will be empty)
	var sttService *service.STTService
	if cfg.STT.Provider != "" {
		sttSvc, err := service.NewSTTService(&cfg.STT)
		if err != nil {
			logger.Warnf("STT service init failed (transcription disabled): %v", err)
		} else {
			sttService = sttSvc
			defer sttService.Close()
			logger.Infof("STT service initialized with provider: %s", cfg.STT.Provider)
		}
	} else {
		logger.Warn("STT provider not configured (transcription disabled)")
	}

	// MemoryTaskService — shared business logic for memory/task operations
	memTaskSvc := service.NewMemoryTaskService(memRepo, taskRepo, vectorStoreSvc, logger)

	// ToolRegistry — register all agent tools with scope-based filtering
	toolRegistry := service.NewToolRegistry()
	// Chat-scoped tools (available to ChatAgent via function calling)
	toolRegistry.Register(tools.NewMemoryCreateTool(memTaskSvc), "chat")
	toolRegistry.Register(tools.NewMemorySearchTool(memTaskSvc), "chat")
	toolRegistry.Register(tools.NewMemoryDeleteTool(memTaskSvc), "chat")
	toolRegistry.Register(tools.NewTaskCreateTool(memTaskSvc), "chat")
	toolRegistry.Register(tools.NewTaskListTool(memTaskSvc), "chat")
	toolRegistry.Register(tools.NewTaskUpdateTool(memTaskSvc), "chat")
	toolRegistry.Register(tools.NewTaskDeleteTool(memTaskSvc), "chat")
	// Extract-scoped tools (available for batch operations from ExtractAgent)
	toolRegistry.Register(tools.NewMemoryCreateBatchTool(memTaskSvc), "extract")
	toolRegistry.Register(tools.NewTaskCreateBatchTool(memTaskSvc), "extract")
	logger.Info("tool registry initialized with memory and task tools")

	audioSvc := service.NewAudioService(convRepo, memTaskSvc, llmProvider, vectorStoreSvc, storageClient, cfg, logger)
	chatSvc := service.NewChatService(chatRepo, convRepo, llmProvider, vectorStoreSvc, toolRegistry, logger)

	// Typesense search
	tsClient := search.NewClient(&cfg.Typesense)
	if err := tsClient.EnsureCollections(); err != nil {
		logger.Warnf("typesense collection init failed (search disabled): %v", err)
	} else {
		logger.Info("typesense collections ready")
	}
	searchSvc := service.NewSearchService(tsClient, convRepo, memRepo)

	// Initialize per-provider LLM configs with the active config from .env
	if cfg.LLMProviders == nil {
		cfg.LLMProviders = make(config.LLMProviderConfigs)
	}
	if cfg.LLM.Provider != "" {
		cfg.LLMProviders[cfg.LLM.Provider] = cfg.LLM
	}

	// Ensure default admin user exists
	if err := authSvc.EnsureAdminUser(); err != nil {
		logger.Warnf("failed to ensure admin user: %v", err)
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	convHandler := handler.NewConversationHandler(convRepo, storageClient)
	memHandler := handler.NewMemoryHandler(memRepo)
	taskHandler := handler.NewTaskHandler(taskRepo)
	chatHandler := handler.NewChatHandler(chatSvc)
	extHandler := handler.NewExtensionHandler(extRepo)
	searchHandler := handler.NewSearchHandler(searchSvc)
	settingsHandler := handler.NewSettingsHandler(cfg)
	wsHandler := handler.NewWSHandler(jwtMgr, audioSvc, sttService)

	// Gin router
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(logger))
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/signin", authHandler.SignIn)
		auth.POST("/signout", authHandler.SignOut)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.Auth(jwtMgr))
	protected.Use(middleware.RateLimit(rdb, 100, 60*time.Second)) // 100 req/min

	// Auth (protected)
	protected.POST("/auth/password", authHandler.ChangePassword)
	protected.POST("/auth/username", authHandler.ChangeUsername)

	// Conversations
	protected.GET("/conversations", convHandler.List)
	protected.GET("/conversations/:id", convHandler.Get)
	protected.GET("/conversations/:id/audio", convHandler.StreamAudio)
	protected.DELETE("/conversations/:id", convHandler.Delete)

	// Memories
	protected.GET("/memories", memHandler.List)
	protected.GET("/memories/:id", memHandler.Get)
	protected.PUT("/memories/:id", memHandler.Update)
	protected.DELETE("/memories/:id", memHandler.Delete)

	// Tasks
	protected.GET("/tasks", taskHandler.List)
	protected.POST("/tasks", taskHandler.Create)
	protected.PUT("/tasks/:id", taskHandler.Update)
	protected.DELETE("/tasks/:id", taskHandler.Delete)

	// Extensions
	protected.GET("/extensions", extHandler.List)
	protected.GET("/extensions/:id", extHandler.Get)
	protected.POST("/extensions", extHandler.Create)
	protected.PUT("/extensions/:id", extHandler.Update)
	protected.DELETE("/extensions/:id", extHandler.Delete)

	// Chat sessions & messages
	protected.GET("/chat/sessions", chatHandler.ListSessions)
	protected.POST("/chat/sessions", chatHandler.CreateSession)
	protected.PUT("/chat/sessions/:id", chatHandler.UpdateSession)
	protected.DELETE("/chat/sessions/:id", chatHandler.DeleteSession)
	protected.GET("/chat/sessions/:id/messages", chatHandler.Messages)
	protected.POST("/chat/sessions/:id/messages", chatHandler.Send)
	protected.POST("/chat/sessions/:id/messages/stream", chatHandler.SendStream)

	// Search
	protected.GET("/search", searchHandler.Search)
	protected.POST("/search/reindex", searchHandler.Reindex)

	// Settings
	protected.GET("/settings/llm", settingsHandler.GetLLMConfig)
	protected.PUT("/settings/llm", settingsHandler.UpdateLLMConfig)
	protected.GET("/settings/cloud", settingsHandler.GetCloudConfig)
	protected.PUT("/settings/cloud", settingsHandler.UpdateCloudConfig)

	// WebSocket (auth via query param)
	v1.GET("/ws/audio", wsHandler.AudioWS)

	addr := fmt.Sprintf(":%s", cfg.App.Port)
	logger.Infof("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migrations source: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrations driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}