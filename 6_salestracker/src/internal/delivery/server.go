package delivery

import (
	"context"
	"salestracker/src/internal/models"

	"github.com/kxddry/wbf/ginext"
)

type Service interface {
	Post(ctx context.Context, req models.PostRequest) (models.PostResponse, error)
}

// Server is a struct that contains the router.
type Server struct {
	r   *ginext.Engine
	svc Service
}
