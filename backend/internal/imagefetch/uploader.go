package imagefetch

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const presignTTL = time.Hour

type UploaderConfig struct {
	Endpoint  string // internal Docker hostname, e.g. "minio:9000"
	PublicURL string // external URL for presigned URLs, e.g. "https://minio.dev.local"
	AccessKey string
	SecretKey string
	Bucket    string
}

type Uploader struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewUploader(ctx context.Context, cfg UploaderConfig) (*Uploader, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false, // TLS terminated by Traefik; plain HTTP on internal network
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("make bucket: %w", err)
		}
	}

	return &Uploader{client: client, bucket: cfg.Bucket, publicURL: cfg.PublicURL}, nil
}

// Upload stores img in MinIO and returns the object key.
func (u *Uploader) Upload(ctx context.Context, optionID uuid.UUID, img FetchedImage) (string, error) {
	ext := extensionForContentType(img.ContentType)
	key := fmt.Sprintf("options/%s/%s%s", optionID, uuid.New(), ext)

	_, err := u.client.PutObject(ctx, u.bucket, key,
		bytes.NewReader(img.Bytes), int64(len(img.Bytes)),
		minio.PutObjectOptions{ContentType: img.ContentType},
	)
	if err != nil {
		return "", fmt.Errorf("minio put %s: %w", key, err)
	}
	return key, nil
}

// PresignURL returns a time-limited URL for the given key, with the host
// rewritten to PublicURL so browsers receive HTTPS via Traefik.
func (u *Uploader) PresignURL(ctx context.Context, key string) (string, error) {
	raw, err := u.client.PresignedGetObject(ctx, u.bucket, key, presignTTL, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign %s: %w", key, err)
	}
	// Replace internal host with external public URL so browser gets HTTPS.
	pub, _ := url.Parse(u.publicURL)
	raw.Scheme = pub.Scheme
	raw.Host = pub.Host
	return raw.String(), nil
}

// Delete removes an object from MinIO.
func (u *Uploader) Delete(ctx context.Context, key string) error {
	return u.client.RemoveObject(ctx, u.bucket, key, minio.RemoveObjectOptions{})
}

// ListObjects returns all objects in the bucket sorted by LastModified (oldest first).
func (u *Uploader) ListObjects(ctx context.Context) ([]minio.ObjectInfo, error) {
	var objs []minio.ObjectInfo
	for obj := range u.client.ListObjects(ctx, u.bucket, minio.ListObjectsOptions{Recursive: true}) {
		if obj.Err != nil {
			return nil, obj.Err
		}
		objs = append(objs, obj)
	}
	// Sort oldest first for LRU eviction
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].LastModified.Before(objs[j].LastModified)
	})
	return objs, nil
}

func extensionForContentType(ct string) string {
	switch strings.Split(ct, ";")[0] {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return filepath.Ext(ct)
	}
}
