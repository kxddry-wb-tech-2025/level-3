package handlers

import (
	"context"
	"delayed-notifier/internal/models"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateNotification(ctx context.Context, db Storage) func(*gin.Context) {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var note models.Notification

		if err = json.Unmarshal(body, &note); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if note.ID != 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id must not be attached"})
			return
		}

		id, err := db.Add(ctx, &note)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}
