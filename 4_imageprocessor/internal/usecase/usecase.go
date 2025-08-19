package usecase

import (
	"context"
	"fmt"
	"image-processor/internal/domain"
	"time"
)

type FileStorage interface {
	Upload(ctx context.Context, file *domain.File) error
	GetURL(ctx context.Context, fileName string) (string, error)
	Get(ctx context.Context, fileName string) (*domain.File, error)
	Delete(ctx context.Context, fileName string) error
}

type StatusStorage interface {
	AddFile(ctx context.Context, file *domain.File) error
	UpdateStatus(ctx context.Context, fileName string, status string) error
	GetStatus(ctx context.Context, id string) (string, error)
	GetFileName(ctx context.Context, id string) (string, error)
	DeleteFile(ctx context.Context, id string) error
}

type TaskSender interface {
	SendTask(ctx context.Context, task *domain.Task) error
}

type Handler struct {
	fs FileStorage
	ss StatusStorage
	ts TaskSender
}

func New(fs FileStorage, ss StatusStorage, ts TaskSender) *Handler {
	return &Handler{fs: fs, ss: ss, ts: ts}
}

func (h *Handler) UploadImage(ctx context.Context, file *domain.File) error {
	const op = "usecase.UploadImage"

	if err := h.fs.Upload(ctx, file); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.ss.AddFile(ctx, file); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.ts.SendTask(ctx, &domain.Task{
		FileName:  file.Name,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (h *Handler) GetImage(ctx context.Context, id string) (*domain.Image, error) {
	const op = "usecase.GetImage"

	status, err := h.ss.GetStatus(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if status != domain.StatusCompleted {
		return &domain.Image{
			Status: status,
		}, nil

	}

	fileName, err := h.ss.GetFileName(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	url, err := h.fs.GetURL(ctx, fileName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.Image{
		URL:    url,
		Status: status,
	}, nil
}

func (h *Handler) DeleteImage(ctx context.Context, id string) error {
	const op = "usecase.DeleteImage"

	if err := h.ss.DeleteFile(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.fs.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
