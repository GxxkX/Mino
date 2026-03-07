package vectordb

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/mino/backend/internal/config"
	"github.com/sirupsen/logrus"
)

const (
	// Vector dimension — matches OpenAI / Zhipu embedding models.
	// Zhipu embedding-3 outputs 2048-dim; OpenAI ada-002 outputs 1536-dim.
	// We use 1024 as a reasonable default that works with most models.
	// Override via EmbeddingDim if needed.
	DefaultEmbeddingDim = 1024

	// Field names used in Milvus collections.
	FieldID        = "id"
	FieldUserID    = "user_id"
	FieldSourceID  = "source_id" // conversation_id or memory_id
	FieldText      = "text"
	FieldVector    = "vector"
	FieldCreatedAt = "created_at"
)

// SearchResult holds a single vector search hit.
type SearchResult struct {
	SourceID string
	UserID   string
	Text     string
	Score    float32
}

// Client wraps the Milvus SDK client with collection management helpers.
type Client struct {
	milvus       client.Client
	cfg          *config.MilvusConfig
	logger       *logrus.Logger
	embeddingDim int
}

// NewClient connects to Milvus and returns a ready-to-use client.
func NewClient(cfg *config.MilvusConfig, logger *logrus.Logger) (*Client, error) {
	if logger == nil {
		logger = logrus.New()
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	ctx := context.Background()
	c, err := client.NewClient(ctx, client.Config{
		Address:  addr,
		Username: cfg.User,
		Password: cfg.Password,
		DBName:   cfg.DBName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Milvus at %s: %w", addr, err)
	}

	logger.Infof("connected to Milvus at %s (db=%s)", addr, cfg.DBName)

	return &Client{
		milvus:       c,
		cfg:          cfg,
		logger:       logger,
		embeddingDim: DefaultEmbeddingDim,
	}, nil
}

// SetEmbeddingDim overrides the default vector dimension.
func (c *Client) SetEmbeddingDim(dim int) {
	c.embeddingDim = dim
}

// Close releases the Milvus connection.
func (c *Client) Close() error {
	return c.milvus.Close()
}

// EnsureCollections creates the conversations and memories collections if they don't exist,
// then loads them into memory for searching.
func (c *Client) EnsureCollections(ctx context.Context) error {
	collections := []string{c.cfg.ConversationsCollection, c.cfg.MemoriesCollection}
	for _, name := range collections {
		exists, err := c.milvus.HasCollection(ctx, name)
		if err != nil {
			return fmt.Errorf("check collection %s: %w", name, err)
		}
		if !exists {
			if err := c.createCollection(ctx, name); err != nil {
				return err
			}
			c.logger.Infof("created Milvus collection: %s", name)
		}

		// Load collection into memory for search
		if err := c.milvus.LoadCollection(ctx, name, false); err != nil {
			c.logger.Warnf("load collection %s: %v", name, err)
		}
	}
	return nil
}

// createCollection builds a collection schema with standard fields.
func (c *Client) createCollection(ctx context.Context, name string) error {
	schema := &entity.Schema{
		CollectionName: name,
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       FieldID,
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     FieldUserID,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "64",
				},
			},
			{
				Name:     FieldSourceID,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "64",
				},
			},
			{
				Name:     FieldText,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "8192",
				},
			},
			{
				Name:     FieldVector,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: fmt.Sprintf("%d", c.embeddingDim),
				},
			},
		},
	}

	if err := c.milvus.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("create collection %s: %w", name, err)
	}

	// Create IVF_FLAT index on the vector field for fast ANN search.
	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 128)
	if err != nil {
		return fmt.Errorf("create index params: %w", err)
	}
	if err := c.milvus.CreateIndex(ctx, name, FieldVector, idx, false); err != nil {
		return fmt.Errorf("create index on %s: %w", name, err)
	}

	return nil
}

// Insert adds a vector with metadata to the specified collection.
func (c *Client) Insert(ctx context.Context, collection, userID, sourceID, text string, vector []float32) error {
	userIDs := []string{userID}
	sourceIDs := []string{sourceID}
	texts := []string{text}
	vectors := [][]float32{vector}

	_, err := c.milvus.Insert(ctx, collection, "",
		entity.NewColumnVarChar(FieldUserID, userIDs),
		entity.NewColumnVarChar(FieldSourceID, sourceIDs),
		entity.NewColumnVarChar(FieldText, texts),
		entity.NewColumnFloatVector(FieldVector, c.embeddingDim, vectors),
	)
	if err != nil {
		return fmt.Errorf("insert into %s: %w", collection, err)
	}

	// Flush to ensure data is persisted and searchable.
	if err := c.milvus.Flush(ctx, collection, false); err != nil {
		c.logger.Warnf("flush %s: %v", collection, err)
	}

	return nil
}

// InsertBatch adds multiple vectors to the specified collection.
func (c *Client) InsertBatch(ctx context.Context, collection string, userIDs, sourceIDs, texts []string, vectors [][]float32) error {
	if len(userIDs) == 0 {
		return nil
	}

	_, err := c.milvus.Insert(ctx, collection, "",
		entity.NewColumnVarChar(FieldUserID, userIDs),
		entity.NewColumnVarChar(FieldSourceID, sourceIDs),
		entity.NewColumnVarChar(FieldText, texts),
		entity.NewColumnFloatVector(FieldVector, c.embeddingDim, vectors),
	)
	if err != nil {
		return fmt.Errorf("batch insert into %s: %w", collection, err)
	}

	if err := c.milvus.Flush(ctx, collection, false); err != nil {
		c.logger.Warnf("flush %s: %v", collection, err)
	}

	return nil
}

// Search performs an ANN search on the specified collection, filtered by user_id.
func (c *Client) Search(ctx context.Context, collection, userID string, queryVector []float32, topK int) ([]SearchResult, error) {
	sp, err := entity.NewIndexIvfFlatSearchParam(16)
	if err != nil {
		return nil, fmt.Errorf("create search params: %w", err)
	}

	filter := fmt.Sprintf(`user_id == "%s"`, userID)

	results, err := c.milvus.Search(
		ctx,
		collection,
		nil, // partitions
		filter,
		[]string{FieldUserID, FieldSourceID, FieldText},
		[]entity.Vector{entity.FloatVector(queryVector)},
		FieldVector,
		entity.COSINE,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("search %s: %w", collection, err)
	}

	var hits []SearchResult
	for _, result := range results {
		var userIDCol *entity.ColumnVarChar
		var sourceIDCol *entity.ColumnVarChar
		var textCol *entity.ColumnVarChar

		for _, field := range result.Fields {
			switch field.Name() {
			case FieldUserID:
				userIDCol = field.(*entity.ColumnVarChar)
			case FieldSourceID:
				sourceIDCol = field.(*entity.ColumnVarChar)
			case FieldText:
				textCol = field.(*entity.ColumnVarChar)
			}
		}

		for i := 0; i < result.ResultCount; i++ {
			hit := SearchResult{
				Score: result.Scores[i],
			}
			if userIDCol != nil {
				v, _ := userIDCol.ValueByIdx(i)
				hit.UserID = v
			}
			if sourceIDCol != nil {
				v, _ := sourceIDCol.ValueByIdx(i)
				hit.SourceID = v
			}
			if textCol != nil {
				v, _ := textCol.ValueByIdx(i)
				hit.Text = v
			}
			hits = append(hits, hit)
		}
	}

	return hits, nil
}

// DeleteBySourceID removes all vectors matching a source_id in the given collection.
func (c *Client) DeleteBySourceID(ctx context.Context, collection, sourceID string) error {
	expr := fmt.Sprintf(`source_id == "%s"`, sourceID)
	if err := c.milvus.Delete(ctx, collection, "", expr); err != nil {
		return fmt.Errorf("delete from %s where %s: %w", collection, expr, err)
	}
	return nil
}

// ConversationsCollection returns the configured collection name for conversations.
func (c *Client) ConversationsCollection() string {
	return c.cfg.ConversationsCollection
}

// MemoriesCollection returns the configured collection name for memories.
func (c *Client) MemoriesCollection() string {
	return c.cfg.MemoriesCollection
}
