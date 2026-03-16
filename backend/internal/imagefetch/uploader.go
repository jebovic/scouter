package imagefetch

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// publicReadPolicy returns an S3 bucket policy that allows anonymous GET on all objects.
func publicReadPolicy(bucket string) string {
	return `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucket + `/*"]}]}`
}

type UploaderConfig struct {
	Endpoint  string // internal Docker hostname, e.g. "minio:9000"
	PublicURL string // external URL served via Traefik, e.g. "https://minio.dev.local"
	AccessKey string
	SecretKey string
	Bucket    string
}

type Uploader struct {
	client *minio.Client
	bucket string
	pubURL *url.URL
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

	// Allow anonymous reads so browsers can load images directly via plain URLs.
	if err := client.SetBucketPolicy(ctx, cfg.Bucket, publicReadPolicy(cfg.Bucket)); err != nil {
		return nil, fmt.Errorf("set bucket policy: %w", err)
	}

	pubURL, err := url.Parse(cfg.PublicURL)
	if err != nil {
		return nil, fmt.Errorf("parse public URL: %w", err)
	}

	return &Uploader{client: client, bucket: cfg.Bucket, pubURL: pubURL}, nil
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

// PresignURL returns a plain public URL for the given key.
// The bucket is publicly readable, so no signature is needed.
func (u *Uploader) PresignURL(_ context.Context, key string) (string, error) {
	return u.pubURL.String() + "/" + u.bucket + "/" + key, nil
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
			return nil, fmt.Errorf("list objects: %w", obj.Err)
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
		return ""
	}
}
