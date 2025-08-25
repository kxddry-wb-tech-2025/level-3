package delivery

import (
	"context"
	"net/http"
	"salestracker/src/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

type Service interface {
	Post(ctx context.Context, req models.PostRequest) (models.PostResponse, error)
}

// Server is a struct that contains the router.
type Server struct {
	r   *ginext.Engine
	svc Service
}

func New(svc Service) *Server {
	r := ginext.New()
	srv := &Server{
		r:   r,
		svc: svc,
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

		res, err := s.svc.Post(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

}
