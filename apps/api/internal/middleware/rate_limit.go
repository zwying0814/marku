package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"marku/api/internal/config"
	"marku/api/internal/response"
)

func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	store := newLimiterStore(rate.Limit(cfg.RequestsPerSecond), cfg.Burst)

	return func(c *gin.Context) {
		if !store.allow(c.ClientIP()) {
			c.Header("Retry-After", "1")
			c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Burst))
			c.Header("X-RateLimit-Remaining", "0")
			response.Abort(c, response.New(http.StatusTooManyRequests, response.CodeRateLimited, "too many requests"))
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Burst))
		c.Next()
	}
}

type limiterStore struct {
	mu          sync.Mutex
	clients     map[string]*limiterEntry
	limit       rate.Limit
	burst       int
	lastCleanup time.Time
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newLimiterStore(limit rate.Limit, burst int) *limiterStore {
	return &limiterStore{
		clients:     make(map[string]*limiterEntry),
		limit:       limit,
		burst:       burst,
		lastCleanup: time.Now(),
	}
}

func (s *limiterStore) allow(key string) bool {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupLocked(now)
	entry, ok := s.clients[key]
	if !ok {
		entry = &limiterEntry{limiter: rate.NewLimiter(s.limit, s.burst)}
		s.clients[key] = entry
	}
	entry.lastSeen = now
	return entry.limiter.Allow()
}

func (s *limiterStore) cleanupLocked(now time.Time) {
	if now.Sub(s.lastCleanup) < 10*time.Minute && len(s.clients) < 10000 {
		return
	}

	for key, entry := range s.clients {
		if now.Sub(entry.lastSeen) > 10*time.Minute {
			delete(s.clients, key)
		}
	}
	s.lastCleanup = now
}
