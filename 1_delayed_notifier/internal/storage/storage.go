package storage

import (
	"time"

	"github.com/kxddry/wbf/retry"
)

// Strategy outlines the way different components should do backoff.
// Those numbers were not selected carefully, I just chose them at random.
var Strategy = retry.Strategy{Attempts: 3, Backoff: 2, Delay: 5 * time.Second}
