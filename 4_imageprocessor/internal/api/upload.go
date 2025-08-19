package api

import (
	"fmt"
	"image-processor/internal/domain"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

// uploadImage is the handler for the upload image route.
func (s *Server) uploadImage() gin.HandlerFunc {
	return func(c *ginext.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check for 20MB limit
		if file.Size > 20*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file size is too large"})
			return
		}

		data, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer data.Close()

		buf := make([]byte, 512)
		if _, err := data.Read(buf); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// check for valid image type
		contentType := http.DetectContentType(buf)
		if contentType != "image/jpeg" && contentType != "image/png" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
			return
		}

		// reset the file pointer to the beginning
		if _, err := data.Seek(0, io.SeekStart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id := uuid.New().String()
		fileName := fmt.Sprintf("%s.%s", id, filepath.Ext(file.Filename))

		// upload the file to the storage
		if err = s.h.UploadImage(c.Request.Context(), &domain.File{
			Name:        fileName,
			Data:        data,
			Size:        file.Size,
			ContentType: contentType,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}
