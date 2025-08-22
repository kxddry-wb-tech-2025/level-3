package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

type Server struct {
	r *ginext.Engine
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
