package api

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
	"github.com/kxddry/wbf/zlog"
	"github.com/rs/zerolog"
)

type Usecase interface {
	CreateEvent(ctx context.Context, event domain.CreateEventRequest) domain.CreateEventResponse
	GetEvent(ctx context.Context, eventID string) domain.EventDetailsResponse
	Book(ctx context.Context, eventID string, userID string) domain.BookResponse
	Confirm(ctx context.Context, eventID, bookingID string) domain.ConfirmResponse
}

type Server struct {
	log zerolog.Logger
	r   *ginext.Engine
	uc  Usecase
	v   *validator.Validate
}

func NewServer(uc Usecase) *Server {
	r := ginext.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(gin.ErrorLogger())
	s := &Server{r: r, uc: uc, v: validator.New(), log: zlog.Logger.With().Str("component", "api").Logger()}

	s.registerRoutes()

	return s
}

func (s *Server) Run(addrs ...string) error {
	if len(addrs) == 0 {
		addrs = []string{":8080"}
	}
	s.log.Info().Msgf("starting server on %s", addrs[0])
	return s.r.Run(addrs...)
}

func (s *Server) registerRoutes() {
	r := s.r.Group("/api")
	// health check
	r.GET("/health", func(c *ginext.Context) {
		s.log.Info().Msg("health check")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// use group to add prefix to all routes
	r = r.Group("/events")
	r.POST("", func(c *ginext.Context) {
		var req domain.CreateEventRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			s.log.Error().Err(err).Msg("failed to bind request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := s.v.Struct(req); err != nil {
			s.log.Error().Err(err).Msg("failed to validate request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res := s.uc.CreateEvent(c.Request.Context(), req)
		if res.Error != "" {
			s.log.Error().Err(errors.New(res.Error)).Msg("failed to create event")
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/:id", func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			s.log.Error().Err(err).Msg("invalid event ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		res := s.uc.GetEvent(c.Request.Context(), id)
		if res.Error != "" {
			s.log.Error().Err(errors.New(res.Error)).Msg("failed to get event")
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/:id/book", func(c *ginext.Context) {
		var req domain.BookRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			s.log.Error().Err(err).Msg("failed to bind request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := s.v.Struct(req); err != nil {
			s.log.Error().Err(err).Msg("failed to validate request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		eventID := c.Param("id")
		if _, err := uuid.Parse(eventID); err != nil {
			s.log.Error().Err(err).Msg("invalid event ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		res := s.uc.Book(c.Request.Context(), eventID, req.UserID)
		if res.Error != "" {
			s.log.Error().Err(errors.New(res.Error)).Msg("failed to book event")
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/:id/confirm", func(c *ginext.Context) {
		var req domain.ConfirmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			s.log.Error().Err(err).Msg("failed to bind request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := s.v.Struct(req); err != nil {
			s.log.Error().Err(err).Msg("failed to validate request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		eventID := c.Param("id")
		if _, err := uuid.Parse(eventID); err != nil {
			s.log.Error().Err(err).Msg("invalid event ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		res := s.uc.Confirm(c.Request.Context(), eventID, req.BookingID)
		if res.Error != "" {
			s.log.Error().Err(errors.New(res.Error)).Msg("failed to confirm booking")
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})
}
