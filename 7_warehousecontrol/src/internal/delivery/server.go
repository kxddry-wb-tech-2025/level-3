package delivery

import (
	"context"
	"net/http"
	"warehousecontrol/src/internal/config"
	"warehousecontrol/src/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kxddry/wbf/ginext"
	"github.com/kxddry/wbf/zlog"
	"github.com/rs/zerolog"
)

type Service interface {
	CreateItem(ctx context.Context, req models.PostItemRequest) (models.Item, error)
	GetItems(ctx context.Context) ([]models.Item, error)
	UpdateItem(ctx context.Context, req models.PutItemRequest) (models.Item, error)
	DeleteItem(ctx context.Context, id string) error
}

type Server struct {
	log    zerolog.Logger
	r      *ginext.Engine
	svc    Service
	v      *validator.Validate
	secret string
}

func NewServer(cfg *config.Config) *Server {
	r := ginext.New()
	log := zlog.Logger.With().Str("service", "server").Logger()
	s := &Server{
		log:    log,
		r:      r,
		v:      validator.New(),
		secret: cfg.JWT.Secret,
	}
	s.registerRoutes(cfg)

	return s
}

func (s *Server) registerRoutes(cfg *config.Config) {
	ui := s.r.Group("/ui")

	if cfg.Server.StaticDir != "" {
		ui.StaticFS("/", http.Dir(cfg.Server.StaticDir))
	}

	api := s.r.Group("/api/v1")

	items := api.Group("/items")
	items.POST("", s.postItem())
	items.GET("", s.getItems())
	items.PUT("/:id", s.putItem())
	items.DELETE("/:id", s.deleteItem())

	meta := api.Group("/meta")
	meta.GET("/health", func(c *ginext.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	meta.GET("/jwt", s.getJWT())
}

func (s *Server) Run(address ...string) error {
	return s.r.Run(address...)
}
