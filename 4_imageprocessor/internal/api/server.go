package api

import (
	"context"
	"fmt"
	"image-processor/internal/domain"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

// Handler is the interface that wraps the basic methods for the API.
type Handler interface {
	UploadImage(ctx context.Context, file *domain.File) error
	GetImage(ctx context.Context, id string) (*domain.Image, error)
	DeleteImage(ctx context.Context, id string) error
}

// Server is the struct that contains the engine and the handler.
type Server struct {
	r *ginext.Engine
	h Handler
}

// New creates a new server with the given handler.
func New(h Handler) *Server {
	r := ginext.New()
	return &Server{
		r: r,
		h: h,
	}
}

// Run starts the server.
func (s *Server) Run(addrs ...string) error {
	s.registerRoutes()
	return s.r.Run(addrs...)
}

// registerRoutes registers the routes for the server.
func (s *Server) registerRoutes() {
	s.r.POST("/upload", s.uploadImage())
	s.r.GET("/image/:id", s.getImage())
	s.r.DELETE("/image/:id", s.deleteImage())
}

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

// getImage is the handler for the get image route.
func (s *Server) getImage() gin.HandlerFunc {
	return func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		image, err := s.h.GetImage(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, image)
	}
}

// deleteImage is the handler for the delete image route.
func (s *Server) deleteImage() gin.HandlerFunc {
	return func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := s.h.DeleteImage(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "success", "id": id})
	}
}
