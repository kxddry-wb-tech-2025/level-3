package api

import (
	"comment-tree/internal/domain"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kxddry/wbf/ginext"
)

type Storage interface {
	AddComment(ctx context.Context, comment domain.Comment) error
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
