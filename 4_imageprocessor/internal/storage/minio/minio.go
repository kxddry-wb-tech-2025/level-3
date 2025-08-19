package minio

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

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

func (s *Storage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	id := uuid.New().String()

	fileName := fmt.Sprintf("%s.%s", id, filepath.Ext(file.Filename))

	_, err = s.client.PutObject(ctx, s.bucketName, fileName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", err
	}

	return id, nil
}
