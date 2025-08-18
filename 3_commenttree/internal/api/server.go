package api

import (
	"comment-tree/internal/domain"
	"comment-tree/internal/storage"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

type Storage interface {
	AddComment(ctx context.Context, comment domain.Comment) error
	GetComments(ctx context.Context, parentID string) (domain.CommentTree, error)
	DeleteComments(ctx context.Context, id string) error
}

type Server struct {
	r  *ginext.Engine
	st Storage
}

func New(st Storage) *Server {
	r := ginext.New()

	s := &Server{r: r, st: st}

	s.setRoutes()
	return s
}

func (s *Server) setRoutes() {
	s.r.POST("/comments", s.postComment())
	s.r.GET("/comments", s.getComment())
	s.r.DELETE("/comments/:id", s.deleteComment())
}

func (s *Server) getComment() gin.HandlerFunc {
	return func(c *ginext.Context) {
		parentID := c.Query("parent")
		if _, err := uuid.Parse(parentID); parentID != "" && err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent id"})
			return
		}

		comment, err := s.st.GetComments(c.Request.Context(), parentID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "comments not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, comment)
	}
}

func (s *Server) Run() error {
	return s.r.Run()
}

func (s *Server) postComment() gin.HandlerFunc {
	return func(c *ginext.Context) {
		var req domain.AddCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		comment := domain.Comment{
			ID:        uuid.New().String(),
			Content:   req.Content,
			ParentID:  req.ParentID,
			CreatedAt: time.Now(),
		}

		if err := s.st.AddComment(c.Request.Context(), comment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, comment)
	}
}

func (s *Server) deleteComment() gin.HandlerFunc {
	return func(c *ginext.Context) {
		id := c.Param("id")
		if _, err := uuid.Parse(id); id != "" && err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := s.st.DeleteComments(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "comments deleted"})
	}
}
