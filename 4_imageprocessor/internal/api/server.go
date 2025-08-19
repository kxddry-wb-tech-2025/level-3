package api

import (
	"image-processor/internal/domain"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

type Storage interface {
	Upload(file *multipart.FileHeader) (string, error)
}

type TaskSender interface {
	SendTask(task *domain.Task) error
}

type ImageGetter interface {
	GetImage(id string) (*domain.Image, error)
}

type ImageDeleter interface {
	DeleteImage(id string) error
}

type Server struct {
	r  *ginext.Engine
	ig ImageGetter
	id ImageDeleter
	st Storage
	ts TaskSender
}

func New(r *ginext.Engine, ig ImageGetter, id ImageDeleter, st Storage, ts TaskSender) *Server {
	return &Server{
		r:  r,
		ig: ig,
		id: id,
		st: st,
		ts: ts,
	}
}

func (s *Server) Run() error {
	s.registerRoutes()
	return s.r.Run()
}

func (s *Server) registerRoutes() {
	s.r.POST("/upload", s.uploadImage())
	s.r.GET("/image/:id", s.getImage())
	s.r.DELETE("/image/:id", s.deleteImage())
}

func (s *Server) uploadImage() gin.HandlerFunc {
	return func(c *ginext.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		uuid, err := s.st.Upload(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		task := &domain.Task{
			ID:        uuid,
			Status:    domain.StatusPending,
			CreatedAt: time.Now(),
		}

		if err := s.ts.SendTask(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": uuid, "time": task.CreatedAt, "status": task.Status})
	}
}

func (s *Server) getImage() gin.HandlerFunc {
	return func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		image, err := s.ig.GetImage(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, image)
	}
}

func (s *Server) deleteImage() gin.HandlerFunc {
	return func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := s.id.DeleteImage(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "success", "id": id})
	}
}
