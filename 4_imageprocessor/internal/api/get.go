package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

// getImage is the handler for the get image route.
func (s *Server) getImage() gin.HandlerFunc {
	return func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		image, err := s.h.GetImage(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, image)
	}
}
