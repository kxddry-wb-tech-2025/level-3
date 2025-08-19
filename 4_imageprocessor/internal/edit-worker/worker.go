package editworker

import (
	"context"
	"image-processor/internal/domain"
	"image-processor/internal/helpers"

	"github.com/google/uuid"
	"github.com/kxddry/wbf/zlog"
)

type Editor interface {
	Resize(ctx context.Context, file *domain.File) error
}

type Storage interface {
	GetStatus(ctx context.Context, id string) (string, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	AddNewID(ctx context.Context, fileName string, newID string) error
}

type FileStorage interface {
	Get(ctx context.Context, id string) (*domain.File, error)
	Upload(ctx context.Context, file *domain.File) error
}

type Worker struct {
	editor Editor
	st     Storage
	fs     FileStorage
}

func NewWorker(editor Editor, st Storage, fs FileStorage) *Worker {
	return &Worker{editor: editor, st: st, fs: fs}
}

func (w *Worker) Handle(ctx context.Context, ch <-chan *domain.Task) {
	const op = "editworker.Handle"
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-ch:
			if !ok {
				return
			}
			status, err := w.st.GetStatus(ctx, task.FileName)
			if err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to get status")
				continue
			}

			if string(status) != domain.StatusPending {
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
			if err := w.editor.Resize(ctx, file); err != nil {
				zlog.Logger.Err(err).Str("op", op).Msg("failed to resize file")

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
		}
	}
}
