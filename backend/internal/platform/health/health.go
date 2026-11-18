package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/go-redis/redis/v8"
)

// DBChecker implements database health checks
type DBChecker struct {
	DSN string
}

// Server handles health check endpoints
type Server struct {
	db  *DBChecker
	redis *redis.Client
}

// New creates a new health server
func New(db *DBChecker, redisClient *redis.Client) *Server {
	return &Server{
		db:    db,
		redis: redisClient,
	}
}

// Healthz returns HTTP 200 when service is alive
func (s *Server) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "ok",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Readyz returns HTTP 200 when database is ready, or 503 when not
func (s *Server) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response := map[string]interface{}{
		"status": "ok",
		"checks": map[string]string{
			"db": "ok",
		},
	}

	// Check database
	if s.db != nil {
		conn, err := pgx.Connect(ctx, s.db.DSN)
		if err != nil {
			response["status"] = "degraded"
			response["checks"].(map[string]string)["db"] = "error"
			w.WriteHeader(http.StatusServiceUnavailable)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		conn.Close(ctx)
	}

	// Check Redis if configured
	if s.redis != nil {
		pong, err := s.redis.Ping(ctx).Result()
		if err != nil || pong != "PONG" {
			response["status"] = "degraded"
			w.WriteHeader(http.StatusServiceUnavailable)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}