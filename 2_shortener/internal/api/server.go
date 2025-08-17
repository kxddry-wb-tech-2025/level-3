package api

import (
	"context"
	"shortener/internal/domain"
	"shortener/internal/validator"
	"time"

	"github.com/kxddry/wbf/ginext"
)

type URLStorage interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURL(ctx context.Context, shortCode string) (string, error)
}

type ClickStorage interface {
	SaveClick(ctx context.Context, click domain.Click) error
	GetClicks(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error)
	ClickCount(ctx context.Context, shortCode string) (int64, error)
	UniqueClickCount(ctx context.Context, shortCode string) (int64, error)
	ClicksByDay(ctx context.Context, shortCode string, start, end time.Time) (map[string]int64, error)
	ClicksByMonth(ctx context.Context, shortCode string, start, end time.Time) (map[string]int64, error)
	ClicksByUserAgent(ctx context.Context, shortCode string, start, end time.Time, limit int) (map[string]int64, error)
}

type Server struct {
	g            *ginext.Engine
	addrs        []string
	urlStorage   URLStorage
	clickStorage ClickStorage
	validator    validator.Validator
}

func New(urlStorage URLStorage, clickStorage ClickStorage, validator validator.Validator, addrs ...string) *Server {
	if len(addrs) == 0 {
		addrs = []string{"0.0.0.0:8080"}
	}
	g := ginext.New()

	return &Server{g: g, addrs: addrs, urlStorage: urlStorage, clickStorage: clickStorage, validator: validator}
}

func (s *Server) Run(ctx context.Context) error {
	return s.g.Run(s.addrs...)
}

func (s *Server) RegisterRoutes(r *ginext.Engine) {
	r.POST("/shorten", s.postShorten())
	r.GET("/s/:short_code", s.getShorten())
}
