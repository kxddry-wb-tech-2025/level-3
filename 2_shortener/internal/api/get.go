package api

import (
	"context"
	"errors"
	"net/http"
	"shortener/internal/domain"
	"shortener/internal/storage"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
)

func (s *Server) getShorten(ctx context.Context) func(c *ginext.Context) {
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

		go s.clickStorage.SaveClick(ctx, domain.Click{
			ShortCode: shortCode,
			UserAgent: c.GetHeader("User-Agent"),
			IP:        c.ClientIP(),
			Referer:   c.GetHeader("Referer"),
			Timestamp: time.Now(),
		})

		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func (s *Server) getAnalytics() func(c *ginext.Context) {
	return func(c *ginext.Context) {
		shortCode := c.Param("short_code")
		if shortCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
			return
		}

		var from, to time.Time
		if v := c.Query("from"); v != "" {
			parsed, err := time.Parse(time.DateOnly, v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date; expected YYYY-MM-DD"})
				return
			}
			from = parsed
		}
		if v := c.Query("to"); v != "" {
			parsed, err := time.Parse(time.DateOnly, v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date; expected YYYY-MM-DD"})
				return
			}
			to = parsed
		}

		resp, err := s.clickStorage.Analytics(c.Request.Context(), shortCode, &from, &to, 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}
