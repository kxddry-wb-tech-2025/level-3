package delivery

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"warehousecontrol/src/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

func (s *Server) postItem() gin.HandlerFunc {
	log := s.log.With().Str("handler", "postItem").Logger()
	return func(c *ginext.Context) {
		var req models.PostItemRequest

		userID, ok := c.Get("id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		uid, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}

		role, ok := c.Get("role")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}

		roleEnum, ok := role.(models.Role)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error().Err(err).Msg("failed to bind request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.v.Struct(req); err != nil {
			log.Error().Err(err).Msg("failed to validate request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := s.svc.CreateItem(c.Request.Context(), req, roleEnum, uid)
		if err != nil {
			log.Error().Err(err).Msg("failed to create item")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, item)
	}
}

func (s *Server) getItems() gin.HandlerFunc {
	log := s.log.With().Str("handler", "getItems").Logger()
	return func(c *ginext.Context) {
		items, err := s.svc.GetItems(c.Request.Context())
		if err != nil {
			log.Error().Err(err).Msg("failed to get items")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, items)
	}
}

func (s *Server) putItem() gin.HandlerFunc {
	log := s.log.With().Str("handler", "putItem").Logger()
	return func(c *ginext.Context) {
		var req models.PutItemRequest

		userID, ok := c.Get("id")

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		uid, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}

		role, ok := c.Get("role")

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		roleEnum, ok := role.(models.Role)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error().Err(err).Msg("failed to bind request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.v.Struct(req); err != nil {
			log.Error().Err(err).Msg("failed to validate request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := s.svc.UpdateItem(c.Request.Context(), req, roleEnum, uid)
		if err != nil {
			log.Error().Err(err).Msg("failed to update item")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

func (s *Server) deleteItem() gin.HandlerFunc {
	log := s.log.With().Str("handler", "deleteItem").Logger()
	return func(c *ginext.Context) {
		id := c.Param("id")
		userID, ok := c.Get("id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		uid := userID.(string)

		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: must be uuid"})
			return
		}

		role, ok := c.Get("role")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}

		roleEnum, ok := role.(models.Role)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}

		if err := s.svc.DeleteItem(c.Request.Context(), id, uid, roleEnum); err != nil {
			if errors.Is(err, models.ErrItemNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			log.Error().Err(err).Msg("failed to delete item")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func (s *Server) createJWT() gin.HandlerFunc {
	log := s.log.With().Str("handler", "getJWT").Logger()
	return func(c *ginext.Context) {
		role := c.Param("role")
		roleInt, err := strconv.Atoi(role)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}

		signed, err := s.authSvc.CreateJWT(c.Request.Context(), models.Role(roleInt))
		if err != nil {
			log.Error().Err(err).Msg("failed to create JWT")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"jwt": signed})
	}
}

func (s *Server) getHistory() gin.HandlerFunc {
	log := s.log.With().Str("handler", "getHistory").Logger()
	return func(c *ginext.Context) {

		dateFromStr := c.Query("date_from")
		dateToStr := c.Query("date_to")
		userID := c.Query("user_id")
		itemID := c.Query("item_id")
		action := c.Query("action")
		role := c.Query("role")
		var dateFrom, dateTo time.Time
		var err error
		if dateFromStr != "" {
			dateFrom, err = time.Parse(time.RFC3339, dateFromStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from"})
				return
			}
		}
		if dateToStr != "" {
			dateTo, err = time.Parse(time.RFC3339, dateToStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to"})
				return
			}
		}
		history, err := s.historySvc.GetHistory(c.Request.Context(), userID, itemID, action, dateFrom, dateTo, role)
		if err != nil {
			if errors.Is(err, models.ErrItemNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			log.Error().Err(err).Msg("failed to get history")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, history)
	}
}
