package api

import (
	"errors"
	"net/http"
	"shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

func (s *Server) getShorten() func(c *ginext.Context) {
	return func(c *ginext.Context) {
		shortCode := c.Param("short_code")
		if shortCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
			return
		}
		url, err := s.urlStorage.GetURL(c.Request.Context(), shortCode)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "short code not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}
