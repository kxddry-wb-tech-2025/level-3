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
	CreateItem(ctx context.Context, req models.PostItemRequest, role models.Role, userID string) (*models.Item, error)
	GetItems(ctx context.Context) ([]models.Item, error)
	UpdateItem(ctx context.Context, req models.PutItemRequest, role models.Role, userID string) (*models.Item, error)
	DeleteItem(ctx context.Context, id string, userID string, role models.Role) error
}

type AuthService interface {
	VerifyJWT(ctx context.Context, tokenString string) (role models.Role, userID string, err error)
	CreateJWT(ctx context.Context, role models.Role) (string, error)
}

type Server struct {
	log     zerolog.Logger
	r       *ginext.Engine
	svc     Service
	authSvc AuthService
	v       *validator.Validate
}

func NewServer(cfg *config.Config, svc Service, authSvc AuthService) *Server {
	r := ginext.New()
	log := zlog.Logger.With().Str("service", "server").Logger()
	s := &Server{
		log:     log,
		r:       r,
		svc:     svc,
		authSvc: authSvc,
		v:       validator.New(),
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
	items.Use(s.VerifyJWT)

	items.POST("", s.postItem())
	items.GET("", s.getItems())
	items.PUT("/:id", s.putItem())
	items.DELETE("/:id", s.deleteItem())

	meta := api.Group("/meta")
	meta.GET("/health", func(c *ginext.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	meta.POST("/jwt/:role", s.createJWT())
}

func (s *Server) Run(address ...string) error {
	return s.r.Run(address...)
}
