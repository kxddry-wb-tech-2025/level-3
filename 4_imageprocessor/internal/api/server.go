package api

import (
	"context"
	"image-processor/internal/domain"

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
