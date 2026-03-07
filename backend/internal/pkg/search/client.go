package search

import (
	"context"
	"fmt"
	"log"

	"github.com/mino/backend/internal/config"
	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"
)

const (
	CollectionConversations = "conversations"
	CollectionMemories      = "memories"
)

// Client wraps the Typesense client and provides collection management.
type Client struct {
	ts *typesense.Client
}

// NewClient creates a Typesense client from config.
func NewClient(cfg *config.TypesenseConfig) *Client {
	ts := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)),
		typesense.WithAPIKey(cfg.APIKey),
	)
	return &Client{ts: ts}
}

// EnsureCollections creates or updates the required collections.
func (c *Client) EnsureCollections() error {
	if err := c.ensureCollection(conversationsSchema()); err != nil {
		return fmt.Errorf("conversations collection: %w", err)
	}
	if err := c.ensureCollection(memoriesSchema()); err != nil {
		return fmt.Errorf("memories collection: %w", err)
	}
	return nil
}

func (c *Client) ensureCollection(schema *api.CollectionSchema) error {
	ctx := context.Background()
	_, err := c.ts.Collection(schema.Name).Retrieve(ctx)
	if err == nil {
		return nil // already exists
	}
	_, err = c.ts.Collections().Create(ctx, schema)
	if err != nil {
		return err
	}
	log.Printf("created Typesense collection: %s", schema.Name)
	return nil
}

// Typesense returns the underlying typesense client for direct access.
func (c *Client) Typesense() *typesense.Client {
	return c.ts
}

// --- Collection schemas ---

func ptr[T any](v T) *T { return &v }

func conversationsSchema() *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: CollectionConversations,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "user_id", Type: "string", Facet: ptr(true)},
			{Name: "title", Type: "string", Optional: ptr(true)},
			{Name: "summary", Type: "string", Optional: ptr(true)},
			{Name: "transcript", Type: "string", Optional: ptr(true)},
			{Name: "language", Type: "string", Facet: ptr(true), Optional: ptr(true)},
			{Name: "status", Type: "string", Facet: ptr(true), Optional: ptr(true)},
			{Name: "created_at", Type: "int64", Sort: ptr(true)},
		},
	}
}

func memoriesSchema() *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: CollectionMemories,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "user_id", Type: "string", Facet: ptr(true)},
			{Name: "conversation_id", Type: "string", Optional: ptr(true)},
			{Name: "content", Type: "string"},
			{Name: "category", Type: "string", Facet: ptr(true), Optional: ptr(true)},
			{Name: "importance", Type: "int32", Sort: ptr(true)},
			{Name: "created_at", Type: "int64", Sort: ptr(true)},
		},
	}
}
