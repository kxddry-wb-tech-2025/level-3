package delivery

import (
	"errors"
	"net/http"
	"warehousecontrol/src/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

func (s *Server) postItem() gin.HandlerFunc {
	log := s.log.With().Str("handler", "postItem").Logger()
	return func(c *ginext.Context) {
		var req models.PostItemRequest

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

		item, err := s.svc.CreateItem(c.Request.Context(), req)
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

		item, err := s.svc.UpdateItem(c.Request.Context(), req)
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
		if _, err := uuid.Parse(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: must be uuid"})
			return
		}
		if err := s.svc.DeleteItem(c.Request.Context(), id); err != nil {
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
		if role != "" && role != models.RoleAdmin && role != models.RoleUser && role != models.RoleManager {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		if role == "" {
			role = models.RoleUser
		}

		signed, err := s.authSvc.CreateJWT(c.Request.Context(), role)
		if err != nil {
			log.Error().Err(err).Msg("failed to create JWT")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"jwt": signed})
	}
}
