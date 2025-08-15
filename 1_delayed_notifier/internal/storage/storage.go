package storage

import (
	"time"

	"github.com/wb-go/wbf/retry"
)

var Strategy = retry.Strategy{Attempts: 3, Backoff: 2, Delay: 5 * time.Second}
