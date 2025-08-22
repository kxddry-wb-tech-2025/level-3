package api

import (
	"eventbooker/src/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

type Server struct {
	r *ginext.Engine
}

type Storage interface {
	CreateEvent(event *domain.CreateEventRequest) (string, error)
	Book(eventID string, userID string)
}

func NewServer() *Server {
	r := ginext.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(gin.ErrorLogger())
	s := &Server{r: r}

	s.registerRoutes()

	return s
}

func (s *Server) Run(addrs ...string) error {
	if len(addrs) == 0 {
		addrs = []string{":8080"}
	}
	return s.r.Run(addrs...)
}

func (s *Server) registerRoutes() {}
