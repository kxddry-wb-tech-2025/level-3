package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) DeleteNotification() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "id is required",
			})
			return
		}

		err := s.store.Cancel(c.Copy(), id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
