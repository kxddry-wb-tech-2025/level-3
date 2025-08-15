package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetNotification handles GET requests on /notify/{id}
func (s *Server) GetNotification() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "id is required",
			})
			return
		}

		st, err := s.store.Get(c.Copy(), id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, st)
		return

	}
}
