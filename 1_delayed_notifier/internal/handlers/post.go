package handlers

import (
	"delayed-notifier/internal/models"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) PostNotification() func(*gin.Context) {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var note models.NotificationCreate

		if err = json.Unmarshal(body, &note); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if note.SendAt.IsZero() {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SendAt is zero"})
			return
		}

		if note.Recipient == "" || note.Channel == "" || note.Message == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing recipient or channel or message"})
			return
		}

		id := uuid.New()
		n := models.Notification{
			ID:        id.String(),
			SendAt:    note.SendAt,
			Channel:   note.Channel,
			Recipient: note.Recipient,
			Message:   note.Message,
			Attempt:   0,
		}

		nSt := models.NotificationStatus{
			ID:        id.String(),
			Status:    models.StatusSent,
			UpdatedAt: time.Now(),
		}

		err = s.store.Set(c.Copy(), nSt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err = s.pub.PublishDelayed(c.Copy(), n); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "status": nSt.Status})
	}
}
