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
func RegisterRoutes(r *ginext.Engine, store *redis.Storage) {
	r.POST("/notify", func(c *ginext.Context) {
		var req createReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Channel == "" || req.Recipient == "" || req.Message == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel, recipient and message are required"})
			return
		}
		// Validate telegram recipient: must be exactly 9 digits
		if req.Channel == "telegram" {
			valid := len(req.Recipient) == 9
			if valid {
				for i := 0; i < 9; i++ {
					ch := req.Recipient[i]
					if ch < '0' || ch > '9' {
						valid = false
						break
					}
				}
			}
			if !valid {
				c.JSON(http.StatusBadRequest, gin.H{"error": "telegram recipient must be exactly 9 digits"})
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
		if err := store.CreateNotification(context.Background(), n); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, n)
	})

	r.GET("/notify/:id", func(c *gin.Context) {
		id := c.Param("id")
		n, err := store.GetNotification(context.Background(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if n == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, n)
	})

	r.DELETE("/notify/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := store.CancelNotification(context.Background(), id); err != nil {
			zlog.Logger.Error().Err(err).Str("id", id).Msg("cancel failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
