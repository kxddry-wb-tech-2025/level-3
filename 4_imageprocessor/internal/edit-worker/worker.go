package editworker

import (
	"context"
	"image-processor/internal/domain"
	"image-processor/internal/helpers"

	"github.com/google/uuid"
	"github.com/kxddry/wbf/zlog"
)

// Editor is the interface that wraps the basic methods for the editor.
type Editor interface {
	Resize(file *domain.File) error
	AddWatermark(file *domain.File) error
}

// Storage is the interface that wraps the basic methods for the storage.
type Storage interface {
	GetStatus(ctx context.Context, id string) (string, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	AddNewID(ctx context.Context, fileName string, newID string) error
}

// FileStorage is the interface that wraps the basic methods for the file storage.
type FileStorage interface {
	Get(ctx context.Context, id string) (*domain.File, error)
	Upload(ctx context.Context, file *domain.File) error
}

// Worker is the struct that contains the editor, storage, and file storage.
type Worker struct {
	editor Editor
	st     Storage
	fs     FileStorage
}

// NewWorker creates a new worker with the given editor, storage, and file storage.
func NewWorker(editor Editor, st Storage, fs FileStorage) *Worker {
	return &Worker{editor: editor, st: st, fs: fs}
}

// Handle starts the worker.
func (w *Worker) Handle(ctx context.Context, ch <-chan *domain.KafkaMessage) {
	const op = "editworker.Handle"
	for {
		select {
		case <-ctx.Done():
			return
		case km, ok := <-ch:
			if !ok {
				return
			}
			task := km.Task
			id, _ := helpers.Split(task.FileName)
			status, err := w.st.GetStatus(ctx, id)
			if err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to get status")
				continue
			}

			if string(status) != domain.StatusPending && string(status) != domain.StatusFailed {
				continue
			}

			file, err := w.fs.Get(ctx, task.FileName)
			if err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to get file")
			}

			if err := w.st.UpdateStatus(ctx, task.FileName, domain.StatusRunning); err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to update status")
				continue
			}

			// only resize for now
			if err := w.editor.AddWatermark(file); err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to add watermark")

				// the file is probably corrupted, so we won't try resizing it again
				if err := km.Commit(); err != nil {
					zlog.Logger.Err(err).Str("op", op).Msg("failed to commit")
				}

				// update status to failed
				if err := w.st.UpdateStatus(ctx, task.FileName, domain.StatusFailed); err != nil {
					zlog.Logger.Err(err).Str("op", op).Msg("failed to update status to failed")
				}
				continue
			}
			newUUID := uuid.New().String()
			_, ext := helpers.Split(task.FileName)

			if err := w.st.AddNewID(ctx, task.FileName, newUUID); err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to add new id")
				continue
			}

			if err := w.fs.Upload(ctx, &domain.File{
				Name:        newUUID + "." + ext,
				Data:        file.Data,
				Size:        file.Size,
				ContentType: file.ContentType,
			}); err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to upload file")
				continue
			}

			if err := w.st.UpdateStatus(ctx, task.FileName, domain.StatusCompleted); err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to update status to completed")
			}

			if err := km.Commit(); err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to commit")
			}
		}
	}
}
