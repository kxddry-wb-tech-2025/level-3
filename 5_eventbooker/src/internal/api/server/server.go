package api

import (
	"context"
	"eventbooker/src/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

type Usecase interface {
	CreateEvent(ctx context.Context, event domain.CreateEventRequest) domain.CreateEventResponse
	GetEvent(ctx context.Context, eventID string) domain.EventDetailsResponse
	Book(ctx context.Context, eventID string, userID string) domain.BookResponse
	Confirm(ctx context.Context, eventID, bookingID string) domain.ConfirmResponse
}

type Server struct {
	r  *ginext.Engine
	uc Usecase
}

func NewServer(uc Usecase) *Server {
	r := ginext.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(gin.ErrorLogger())
	s := &Server{r: r, uc: uc}

	s.registerRoutes()

	return s
}

func (s *Server) Run(addrs ...string) error {
	if len(addrs) == 0 {
		addrs = []string{":8080"}
	}
	return s.r.Run(addrs...)
}

func (s *Server) registerRoutes() {
	// use group to add prefix to all routes
	r := s.r.Group("/api/v1")
	r.POST("/events", func(c *ginext.Context) {
		var req domain.CreateEventRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res := s.uc.CreateEvent(c.Request.Context(), req)
		if res.Error != "" {
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/events/:id", func(c *ginext.Context) {
		id := c.Param("id")
		res := s.uc.GetEvent(c.Request.Context(), id)
		if res.Error != "" {
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/events/:id/book", func(c *ginext.Context) {
		var req domain.BookRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		eventID := c.Param("id")
		// validate event ID
		if _, err := uuid.Parse(eventID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		res := s.uc.Book(c.Request.Context(), eventID, req.UserID)
		if res.Error != "" {
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/events/:id/confirm", func(c *ginext.Context) {
		var req domain.ConfirmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		eventID := c.Param("id")
		if _, err := uuid.Parse(eventID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		res := s.uc.Confirm(c.Request.Context(), eventID, req.BookingID)
		if res.Error != "" {
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})
}
