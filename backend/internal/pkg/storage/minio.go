package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/mino/backend/internal/config"
)

const audioBucket = "mino-audio"

// Client wraps the MinIO SDK for object storage operations.
type Client struct {
	mc          *minio.Client
	internalURL string // internal base URL used by the server-side proxy
}

// NewClient initialises a MinIO client and ensures the audio bucket exists.
func NewClient(cfg *config.MinIOConfig) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client init: %w", err)
	}

	// Ensure the audio bucket exists
	ctx := context.Background()
	exists, err := mc.BucketExists(ctx, audioBucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, audioBucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("minio create bucket: %w", err)
		}
	}

	// Build the internal base URL from the endpoint so the web proxy can
	// always reach MinIO directly, regardless of MINIO_PUBLIC_URL.
	scheme := "http"
	if cfg.Secure {
		scheme = "https"
	}
	internalURL := fmt.Sprintf("%s://%s", scheme, cfg.Endpoint)

	return &Client{mc: mc, internalURL: internalURL}, nil
}

// UploadAudio stores an audio blob (e.g. opus/webm) and returns its internal URL.
// The returned URL points to the MinIO endpoint directly so the Next.js proxy
// can fetch it server-side; it is never exposed to the browser as-is.
func (c *Client) UploadAudio(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := c.mc.PutObject(ctx, audioBucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio put object: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", c.internalURL, audioBucket, objectKey), nil
}

// DeleteAudio removes an audio object from the bucket.
// It first resolves the version ID (if versioning is enabled) so the actual
// object data is purged rather than just receiving a delete marker.
func (c *Client) DeleteAudio(ctx context.Context, objectKey string) error {
	// StatObject to get the current version ID (empty string when unversioned).
	info, err := c.mc.StatObject(ctx, audioBucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		// Object may already be gone — treat as success.
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil
		}
		return fmt.Errorf("minio stat object: %w", err)
	}

	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
		VersionID:        info.VersionID,
	}
	if err := c.mc.RemoveObject(ctx, audioBucket, objectKey, opts); err != nil {
		return fmt.Errorf("minio remove object: %w", err)
	}
	return nil
}
