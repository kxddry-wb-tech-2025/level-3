package minio

import (
	"context"
	"fmt"
	"image-processor/internal/domain"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	BaseURL    string
	SSL        bool
}

// Storage is the main storage struct that contains the minio client.
type Storage struct {
	client     *minio.Client
	bucketName string
	baseURL    string
}

// New creates a new storage with the given endpoint, access key, and secret key.
func New(ctx context.Context, cfg Config) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &Storage{client: client, bucketName: cfg.BucketName, baseURL: cfg.BaseURL}, nil
}

func (s *Storage) Upload(ctx context.Context, file domain.File) (string, error) {
	id := uuid.New().String()

	fileName := fmt.Sprintf("%s.%s", id, filepath.Ext(file.ContentType))

	_, err := s.client.PutObject(ctx, s.bucketName, fileName, file.Data, file.Size, minio.PutObjectOptions{
		ContentType: file.ContentType,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, id string) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, id, time.Hour*24, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}
