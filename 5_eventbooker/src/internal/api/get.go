package api

import (
	"github.com/kxddry/wbf/ginext"
)

func (s *Server) getEvent(c *ginext.Context) {
	id := c.Param("id")
	_ = id
}
