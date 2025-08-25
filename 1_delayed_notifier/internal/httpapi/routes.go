package httpapi

import (
	"context"
	"net/http"
	"time"

	"delayed-notifier/internal/models"
	"delayed-notifier/internal/storage/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
	"github.com/kxddry/wbf/zlog"
)

type createReq struct {
	SendAt    *time.Time `json:"send_at"`
	Channel   string     `json:"channel"`
	Recipient string     `json:"recipient"`
	Message   string     `json:"message"`
}

// RegisterRoutes registers HTTP endpoints for creating, querying and cancelling notifications.
func RegisterRoutes(ctx context.Context, r *ginext.Engine, store *redis.Storage) {
	log := zlog.Logger.With().Str("component", "httpapi").Logger()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(gin.ErrorLogger())
	r.POST("/notify", func(c *ginext.Context) {
		var req createReq
		if err := c.BindJSON(&req); err != nil {
			log.Error().Err(err).Msg("bind json failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Channel == "" || req.Recipient == "" || req.Message == "" {
			log.Error().Msg("channel, recipient and message are required")
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel, recipient and message are required"})
			return
		}
		// Validate telegram recipient: must be between 3 and 13 digits
		if req.Channel == "telegram" {
			valid := len(req.Recipient) >= 3 && len(req.Recipient) <= 13
			if valid {
				for i := 0; i < len(req.Recipient); i++ {
					ch := req.Recipient[i]
					if ch < '0' || ch > '9' {
						valid = false
						break
					}
				}
			}
			if !valid {
				log.Error().Msg("telegram recipient must be between 3 and 13 digits")
				c.JSON(http.StatusBadRequest, gin.H{"error": "telegram recipient must be between 3 and 13 digits"})
				return
			}
		}
		id := uuid.NewString()
		now := time.Now().UTC()
		sendAt := now
		if req.SendAt != nil {
			sendAt = req.SendAt.UTC()
		}
		n := &models.Notification{
			ID:        id,
			Channel:   req.Channel,
			Recipient: req.Recipient,
			Message:   req.Message,
			SendAt:    sendAt,
			Status:    models.StatusScheduled,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := store.CreateNotification(ctx, n); err != nil {
			log.Error().Err(err).Msg("create notification failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, n)
	})

	r.GET("/notify/:id", func(c *ginext.Context) {
		id := c.Param("id")
		n, err := store.GetNotification(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("get notification failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if n == nil {
			log.Error().Msg("notification not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, n)
	})

	r.DELETE("/notify/:id", func(c *ginext.Context) {
		id := c.Param("id")
		if err := store.CancelNotification(ctx, id); err != nil {
			log.Error().Err(err).Str("id", id).Msg("cancel failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
