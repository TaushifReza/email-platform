package ratelimit

import "errors"

// ErrCounterUnavailable is returned when rate limiting cannot run.
var ErrCounterUnavailable = errors.New("ratelimit: counter unavailable")
