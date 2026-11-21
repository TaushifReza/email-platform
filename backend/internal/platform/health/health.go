package health

import (
	"context"
)

// Pinger is the small dependency the readiness check needs.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Status is the result of a readiness check.
type Status struct {
	Ready    bool
	Database string
	Redis    string
}

// Service reports liveness and readiness.
type Service struct {
	db    Pinger
	redis Pinger
}

// New creates a health service. Redis is optional and never gates readiness;
// the architecture promises that Redis downtime loses no work.
func New(db Pinger, redis Pinger) *Service {
	return &Service{db: db, redis: redis}
}

// Live reports whether the process is running.
func (s *Service) Live() bool {
	return true
}

// Ready checks the database and reports Redis status informationally.
func (s *Service) Ready(ctx context.Context) Status {
	status := Status{Ready: true, Database: "ok", Redis: "disabled"}

	if s.db == nil || s.db.Ping(ctx) != nil {
		status.Ready = false
		status.Database = "error"
	}

	if s.redis != nil {
		if s.redis.Ping(ctx) != nil {
			status.Redis = "error"
		} else {
			status.Redis = "ok"
		}
	}

	return status
}
