package rustfs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/withoutforget/fshare/internal/config"
)

// Client wraps minio.Client and provides high-level S3 operations.
type Client struct {
	mc     *minio.Client
	bucket string
}

// NewClient creates a new S3/RustFS client and ensures the target bucket exists.
func NewClient(cfg config.S3Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("rustfs: create client: %w", err)
	}

	c := &Client{mc: mc, bucket: cfg.Bucket}

	if err := c.ensureBucket(context.Background(), cfg.Region); err != nil {
		return nil, err
	}

	slog.Info("rustfs connected", "endpoint", cfg.Endpoint, "bucket", cfg.Bucket)
	return c, nil
}

// ensureBucket creates the bucket when it does not already exist.
func (c *Client) ensureBucket(ctx context.Context, region string) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("rustfs: check bucket: %w", err)
	}
	if exists {
		return nil
	}
	if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		return fmt.Errorf("rustfs: make bucket: %w", err)
	}
	slog.Info("rustfs: bucket created", "bucket", c.bucket)
	return nil
}

// Upload streams r into the bucket under objectName.
//   - size: exact byte count, or -1 when unknown (causes buffering on the server side).
//   - contentType: MIME type; pass "" to let the SDK detect it.
func (c *Client) Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) (minio.UploadInfo, error) {
	opts := minio.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}

	info, err := c.mc.PutObject(ctx, c.bucket, objectName, r, size, opts)
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("rustfs: upload %q: %w", objectName, err)
	}
	return info, nil
}

// Download returns a ReadCloser for objectName. The caller must close it.
func (c *Client) Download(ctx context.Context, objectName string) (io.ReadCloser, minio.ObjectInfo, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, minio.ObjectInfo{}, fmt.Errorf("rustfs: get %q: %w", objectName, err)
	}

	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, minio.ObjectInfo{}, fmt.Errorf("rustfs: stat %q: %w", objectName, err)
	}
	return obj, info, nil
}

// Delete removes objectName from the bucket.
func (c *Client) Delete(ctx context.Context, objectName string) error {
	if err := c.mc.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("rustfs: delete %q: %w", objectName, err)
	}
	return nil
}

// PresignedGetURL returns a temporary GET URL for objectName, valid for expires.
// Use this to hand download links directly to clients without proxying bytes.
//
//	link, err := client.PresignedGetURL(ctx, "uploads/photo.jpg", 24*time.Hour)
func (c *Client) PresignedGetURL(ctx context.Context, objectName string, expires time.Duration) (*url.URL, error) {
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, objectName, expires, nil)
	if err != nil {
		return nil, fmt.Errorf("rustfs: presign %q: %w", objectName, err)
	}
	return u, nil
}

// MC exposes the raw *minio.Client for advanced use-cases
// (multipart uploads, bucket policies, etc.).
func (c *Client) MC() *minio.Client { return c.mc }

// Bucket returns the configured bucket name.
func (c *Client) Bucket() string { return c.bucket }
