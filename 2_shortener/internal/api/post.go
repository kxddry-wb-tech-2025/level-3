package api

import (
	"net/http"
	"regexp"
	"shortener/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

var aliasRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

func (s *Server) postShorten() func(c *ginext.Context) {
	return func(c *ginext.Context) {
		var req domain.ShortenRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := s.validator.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Alias != "" && !aliasRe.MatchString(req.Alias) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alias"})
			return
		}
		shortCode, err := s.urlStorage.SaveURL(c.Request.Context(), req.URL, req.Alias != "", req.Alias)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"short_code": shortCode})
	}
}
