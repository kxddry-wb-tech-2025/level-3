package delivery

import (
	"net/http"

	"github.com/kxddry/wbf/ginext"
	"github.com/kxddry/wbf/zlog"
	"github.com/rs/zerolog"
)

type Server struct {
	log zerolog.Logger
	r   *ginext.Engine
}

func NewServer(staticDir string) *Server {
	r := ginext.New()
	log := zlog.Logger.With().Str("service", "server").Logger()
	s := &Server{
		log: log,
		r:   r,
	}

	s.registerRoutes(staticDir)

	return s
}

func (s *Server) registerRoutes(staticDir string) {
	api := s.r.Group("/api/v1")

	if staticDir != "" {
		s.r.StaticFS("/ui", http.Dir(staticDir))
	}

	api.POST("/items", nil)
	api.GET("/items", nil)
	api.PUT("/items/:id", nil)
	api.DELETE("/items/:id", nil)
}

func (s *Server) Run(address ...string) error {
	return s.r.Run(address...)
}
