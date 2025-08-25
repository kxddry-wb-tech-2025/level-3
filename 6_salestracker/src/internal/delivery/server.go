package delivery

import (
	"context"
	"errors"
	"net/http"
	"salestracker/src/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

// Service is an interface that contains the usecase methods.
type Service interface {
	PostItem(ctx context.Context, req models.PostRequest) (models.PostResponse, error)
	GetItems(ctx context.Context) ([]models.Item, error)
	PutItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error)
	DeleteItem(ctx context.Context, id string) error
	GetAnalytics(ctx context.Context, from, to *time.Time) (models.Analytics, error)
}

// Server is a struct that contains the router.
type Server struct {
	r   *ginext.Engine
	svc Service
	v   *validator.Validate
}

// New creates a new server.
func New(svc Service, staticDir string) *Server {
	r := ginext.New()
	v := validator.New()
	srv := &Server{
		r:   r,
		svc: svc,
		v:   v,
	}

	if staticDir != "" {
		r.StaticFS("/ui", http.Dir(staticDir))
	}

	srv.registerRoutes()

	return srv
}

// Run runs the server.
func (s *Server) Run(addrs ...string) error {
	if len(addrs) == 0 {
		addrs = []string{":8080"}
	}
	zlog.Logger.Info().Msgf("starting server on %s", addrs[0])
	return s.r.Run(addrs...)
}

// registerRoutes registers the routes for the server.
func (s *Server) registerRoutes() {
	r := s.r.Group("/api/v1")

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
			var status int
			if errors.Is(err, models.ErrItemNotFound) {
				status = http.StatusNotFound
			} else {
				status = http.StatusInternalServerError
			}

			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/items/:id", func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		err := s.svc.DeleteItem(c.Request.Context(), id)
		if err != nil {
			var status int
			if errors.Is(err, models.ErrItemNotFound) {
				status = http.StatusNotFound
			} else {
				status = http.StatusInternalServerError
			}

			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	})

	r.GET("/analytics", func(c *ginext.Context) {
		var from, to *time.Time
		fromStr := c.Query("from")
		toStr := c.Query("to")
		if fromStr != "" {
			fromTime, err := time.Parse(time.RFC3339, fromStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date"})
				return
			}
			from = &fromTime
		}
		if toStr != "" {
			toTime, err := time.Parse(time.RFC3339, toStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date"})
				return
			}
			to = &toTime
		}

		analytics, err := s.svc.GetAnalytics(c.Request.Context(), from, to)
		if err != nil {
			var status int
			if errors.Is(err, models.ErrItemNotFound) {
				status = http.StatusNotFound
			} else {
				status = http.StatusInternalServerError
			}

			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, analytics)
	})
}
