package api

import (
	"net/http"
	"shortener/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

func (s *Server) postShorten() func(c *ginext.Context) {
	return func(c *ginext.Context) {
		var req domain.ShortenRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := s.validator.URL(req.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		shortCode, err := s.urlStorage.SaveURL(c.Request.Context(), req.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"short_code": shortCode})
	}
}
