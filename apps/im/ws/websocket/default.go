package websocket

import (
	"math"
	"time"
)

const (
	defaultMaxConnectionIdle = time.Duration(math.MaxInt64)
	defaultAckTime           = 30 * time.Second
	defaultConcurrency       = 10
)
