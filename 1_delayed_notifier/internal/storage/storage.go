package storage

import (
	"errors"
	"time"

	"github.com/wb-go/wbf/retry"
)

var (
	ErrNoSubscription = errors.New("no subscription with that id found")
)

var Strategy = retry.Strategy{Attempts: 3, Backoff: 2, Delay: 5 * time.Second}
