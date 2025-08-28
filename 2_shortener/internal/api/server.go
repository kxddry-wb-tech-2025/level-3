package api

import (
	"context"
	"net/http"
	"shortener/internal/domain"
	"shortener/internal/validator"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/ginext"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// URLStorage is the interface for the URL storage.
type URLStorage interface {
	SaveURL(ctx context.Context, url string, withAlias bool, alias string) (string, error)
	GetURL(ctx context.Context, shortCode string) (string, error)
}

// ClickStorage is the interface for the click storage.
type ClickStorage interface {
	SaveClick(ctx context.Context, click domain.Click) error
	GetClicks(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error)
	ClickCount(ctx context.Context, shortCode string) (int64, error)
	UniqueClickCount(ctx context.Context, shortCode string) (int64, error)
	ClicksByDay(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error)
	ClicksByMonth(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error)
	ClicksByUserAgent(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error)
	Analytics(ctx context.Context, shortCode string, from, to *time.Time, topLimit int) (domain.AnalyticsResponse, error)
	ClicksByReferer(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error)
	ClicksByIP(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error)
}

// CacheStorage is the interface for the cache storage.
type CacheStorage interface {
	GetURL(ctx context.Context, shortCode string) (string, error)
	SetURL(ctx context.Context, shortCode, url string, usage int64) error
}

// Server is the server.
type Server struct {
	g            *ginext.Engine
	addrs        []string
	urlStorage   URLStorage
	clickStorage ClickStorage
	validator    validator.Validator
	cache        CacheStorage
}

// New creates a new server.
func New(urlStorage URLStorage, clickStorage ClickStorage, validator validator.Validator, cache CacheStorage, addrs ...string) *Server {
	if len(addrs) == 0 {
		addrs = []string{"0.0.0.0:8080"}
	}
	g := ginext.New()

	_ = g.SetTrustedProxies(nil)

	return &Server{g: g, addrs: addrs, urlStorage: urlStorage, clickStorage: clickStorage, validator: validator, cache: cache}
}

// Run runs the server.
func (s *Server) Run(ctx context.Context) error {
	return s.g.Run(s.addrs...)
}

// RegisterRoutes registers the routes.
func (s *Server) RegisterRoutes(ctx context.Context) {
	// API routes
	s.g.POST("/shorten", s.postShorten(ctx))
	s.g.GET("/s/:short_code", s.getShorten(ctx))
	s.g.GET("/analytics/:short_code", s.getAnalytics())

	s.g.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	s.g.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Static UI routes
	s.g.Static("/ui", "./web")
	s.g.StaticFile("/", "./web/index.html")
}
