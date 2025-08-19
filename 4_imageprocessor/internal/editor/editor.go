package editor

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image-processor/internal/domain"
	"image/draw"
	"io"

	"github.com/disintegration/imaging"
	"github.com/kxddry/wbf/zlog"
)

// Editor is the main editor struct that contains the editor's methods.
type Editor struct {
	watermark image.Image
	margin    int
}

// NewEditor creates a new editor.
func NewEditor(wmPath string, margin int) (*Editor, error) {
	const op = "editor.NewEditor"
	wm, err := imaging.Open(wmPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Editor{watermark: wm, margin: margin}, nil
}

// checkFormat checks if the file is a supported format.
func (e *Editor) checkFormat(file *domain.File) error {
	if _, ok := format[file.ContentType]; !ok {
		return errors.New("unsupported format")
	}
	return nil
}

// encode encodes the image to the specified format.
func (e *Editor) encode(img image.Image, format imaging.Format) (io.ReadSeekCloser, error) {
	buf := bytes.NewBuffer(nil)
	if err := imaging.Encode(buf, img, format); err != nil {
		return nil, err
	}
	return &readSeekCloser{bytes.NewReader(buf.Bytes())}, nil
}

// format is a map of content types to imaging formats.
var format = map[string]imaging.Format{
	"image/jpg": imaging.JPEG, "image/jpeg": imaging.JPEG,
	"image/png": imaging.PNG,
	"image/gif": imaging.GIF,
}

// readSeekCloser is a wrapper around bytes.Reader that implements io.ReadSeekCloser.
type readSeekCloser struct{ *bytes.Reader }

func (r *readSeekCloser) Close() error { return nil }

// Resize resizes the image 2x.
func (e *Editor) Resize(file *domain.File) error {
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

	file.Data, err = e.encode(img, format[file.ContentType])
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// AddWatermark adds a watermark to the image.
func (e *Editor) AddWatermark(file *domain.File) error {
	const op = "editor.AddWatermark"

	if _, ok := format[file.ContentType]; !ok {
		return fmt.Errorf("%s: %w", op, errors.New("unsupported format"))
	}

	img, err := imaging.Decode(file.Data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_ = file.Data.Close()

	x := img.Bounds().Max.X - e.watermark.Bounds().Dx() - e.margin
	y := img.Bounds().Max.Y - e.watermark.Bounds().Dy() - e.margin

	rgba := imaging.Clone(img)
	draw.Draw(rgba, e.watermark.Bounds().Add(image.Pt(x, y)), e.watermark, image.Point{}, draw.Over)

	file.Data, err = e.encode(rgba, format[file.ContentType])
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rc, ok := file.Data.(*readSeekCloser); ok {
		file.Size = rc.Size()
	} else {
		zlog.Logger.Error().Str("op", op).Msg("file.Data is not a *readSeekCloser")
	}

	return nil
}
