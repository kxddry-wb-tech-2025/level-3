package minio

import (
	"context"
	"errors"
	"fmt"
	"image-processor/internal/domain"
	"image-processor/internal/storage"
	"time"

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
	const op = "storage.minio.New"
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return &Storage{client: client, bucketName: cfg.BucketName, baseURL: cfg.BaseURL}, nil
}

func (s *Storage) Upload(ctx context.Context, file *domain.File) error {
	const op = "storage.minio.Upload"
	_, err := s.client.PutObject(ctx, s.bucketName, file.Name, file.Data, file.Size, minio.PutObjectOptions{
		ContentType: file.ContentType,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(ctx context.Context, fileName string) (string, error) {
	const op = "storage.minio.GetURL"
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, fileName, time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return presignedURL.String(), nil
}

func (s *Storage) Get(ctx context.Context, fileName string) (*domain.File, error) {
	const op = "storage.minio.Get"
	object, err := s.client.GetObject(ctx, s.bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		var errResp minio.ErrorResponse
		if errors.As(err, &errResp) {
			if errResp.Code == "NoSuchKey" {
				return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
			}
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	objectInfo, err := object.Stat()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.File{
		Name:        fileName,
		Data:        object,
		ContentType: objectInfo.ContentType,
		Size:        objectInfo.Size,
	}, nil
}

func (s *Storage) Delete(ctx context.Context, fileName string) error {
	const op = "storage.minio.Delete"
	err := s.client.RemoveObject(ctx, s.bucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
