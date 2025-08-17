package api

import (
	"context"
	"errors"
	"net/http"
	"shortener/internal/domain"
	"shortener/internal/storage"
	"strconv"
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
		from := c.Query("from")
		to := c.Query("to")
		topLimit := c.Query("top_limit")
		if from == "" {
			from = time.Now().AddDate(0, -1, 0).Format(time.DateOnly)
		}
		if to == "" {
			to = time.Now().Format(time.DateOnly)
		}
		if topLimit == "" {
			topLimit = "100"
		}
		fromTime, err := time.Parse(time.DateOnly, from)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date"})
			return
		}
		toTime, err := time.Parse(time.DateOnly, to)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date"})
			return
		}
		topLimitInt, err := strconv.Atoi(topLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid top limit"})
			return
		}
		analytics, err := s.clickStorage.Analytics(c.Request.Context(), shortCode, fromTime, toTime, topLimitInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, analytics)
	}
}
