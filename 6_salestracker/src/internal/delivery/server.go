package delivery

import (
	"context"
	"net/http"
	"salestracker/src/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

type Service interface {
	PostItem(ctx context.Context, req models.PostRequest) (models.PostResponse, error)
	GetItems(ctx context.Context) ([]models.Item, error)
	PutItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error)
}

// Server is a struct that contains the router.
type Server struct {
	r   *ginext.Engine
	svc Service
	v   *validator.Validate
}

func New(svc Service) *Server {
	r := ginext.New()
	v := validator.New()
	srv := &Server{
		r:   r,
		svc: svc,
		v:   v,
	}

	srv.registerRoutes()

	return srv
}

func (s *Server) registerRoutes() {
	r := s.r.Group("/api")

	r.POST("/items", func(c *ginext.Context) {
		var req models.PostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := s.svc.PostItem(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	r.GET("/items", func(c *ginext.Context) {
		items, err := s.svc.GetItems(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, items)
	})

	r.PUT("/items/:id", func(c *ginext.Context) {
		var req models.PutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.v.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// validate uuid
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		res, err := s.svc.PutItem(c.Request.Context(), id, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

}
