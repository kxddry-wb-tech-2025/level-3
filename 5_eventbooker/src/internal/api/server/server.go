package api

import (
	"eventbooker/src/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

type Usecase interface {
	CreateEvent(event *domain.CreateEventRequest) domain.CreateEventResponse
	Book(req *domain.BookRequest) domain.BookResponse
	Confirm(req *domain.ConfirmRequest) domain.ConfirmResponse
	GetEvent(id string) domain.EventDetailsResponse
}

type Server struct {
	r  *ginext.Engine
	uc Usecase
}

type Storage interface {
	CreateEvent(event *domain.CreateEventRequest) (string, error)
	Book(eventID string, userID string)
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
	r := s.r
	r.POST("/events", func(c *ginext.Context) {
		var req domain.CreateEventRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res := s.uc.CreateEvent(&req)
		if res.Error != "" {
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/events/:id", func(c *ginext.Context) {
		id := c.Param("id")
		res := s.uc.GetEvent(id)
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
		res := s.uc.Book(&req)
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
		res := s.uc.Confirm(&req)
		if res.Error != "" {
			c.JSON(http.StatusBadRequest, res)
			return
		}
		c.JSON(http.StatusOK, res)
	})
}
