package editor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image-processor/internal/domain"

	"github.com/disintegration/imaging"
)

// Editor is the main editor struct that contains the editor's methods.
type Editor struct{}

// NewEditor creates a new editor.
func NewEditor() *Editor {
	return &Editor{}
}

// format is a map of content types to imaging formats.
var format = map[string]imaging.Format{
	"jpg": imaging.JPEG, "jpeg": imaging.JPEG,
	"png": imaging.PNG,
	"gif": imaging.GIF,
}

// readSeekCloser is a wrapper around bytes.Reader that implements io.ReadSeekCloser.
type readSeekCloser struct{ *bytes.Reader }

func (r *readSeekCloser) Close() error { return nil }

// Resize resizes the image 2x.
func (e *Editor) Resize(ctx context.Context, file *domain.File) error {
	const op = "editor.Resize"

	if _, ok := format[file.ContentType]; !ok {
		return fmt.Errorf("%s: %w", op, errors.New("unsupported format"))
	}

	img, err := imaging.Decode(file.Data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_ = file.Data.Close()

	img = imaging.Resize(img, 2*img.Bounds().Dx(), 2*img.Bounds().Dy(), imaging.Lanczos)

	buf := bytes.NewBuffer(nil)
	if err := imaging.Encode(buf, img, format[file.ContentType]); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	file.Data = &readSeekCloser{bytes.NewReader(buf.Bytes())}

	return nil
}
