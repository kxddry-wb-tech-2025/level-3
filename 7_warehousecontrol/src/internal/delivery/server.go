package delivery

import (
	"context"
	"net/http"
	"time"
	"warehousecontrol/src/internal/config"
	"warehousecontrol/src/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kxddry/wbf/ginext"
	"github.com/kxddry/wbf/zlog"
	"github.com/rs/zerolog"
)

// Service is the interface for the service.
type Service interface {
	CreateItem(ctx context.Context, req models.PostItemRequest, role models.Role, userID string) (*models.Item, error)
	GetItems(ctx context.Context) ([]models.Item, error)
	UpdateItem(ctx context.Context, req models.PutItemRequest, role models.Role, userID string) (*models.Item, error)
	DeleteItem(ctx context.Context, id string, userID string, role models.Role) error
}

// AuthService is the interface for the auth service.
type AuthService interface {
	VerifyJWT(ctx context.Context, tokenString string) (role models.Role, userID string, err error)
	CreateJWT(ctx context.Context, role models.Role) (string, error)
}

// HistoryService is the interface for the history service.
type HistoryService interface {
	GetHistory(ctx context.Context, role models.Role, filterByUserID string,
		filterByItemID string, filterByAction string, filterDateFrom time.Time, filterDateTo time.Time,
		filterByUserRole string, limit, offset int64) ([]models.HistoryEntry, error)
}

// Server is the server for the service.
type Server struct {
	log        zerolog.Logger
	r          *ginext.Engine
	svc        Service
	authSvc    AuthService
	historySvc HistoryService
	v          *validator.Validate
}

// NewServer creates a new server.
func NewServer(cfg *config.Config, svc Service, authSvc AuthService, historySvc HistoryService) *Server {
	r := ginext.New()
	log := zlog.Logger.With().Str("service", "server").Logger()
	s := &Server{
		log:        log,
		r:          r,
		svc:        svc,
		authSvc:    authSvc,
		historySvc: historySvc,
		v:          validator.New(),
	}
	s.registerRoutes(cfg)

	return s
}

// registerRoutes registers the routes for the server.
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
	history := meta.Group("/history")
	history.Use(s.VerifyJWT)
	history.GET("", s.getHistory())
}

// Run runs the server.
func (s *Server) Run(address ...string) error {
	return s.r.Run(address...)
}
