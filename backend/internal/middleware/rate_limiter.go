package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

/**
 * ipLimiter holds a rate limiter and the last time it was seen.
 * The lastSeen field lets us clean up stale entries from the map.
*/
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterStore manages per-IP limiters safely across goroutines.
type RateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit // tokens per second
	b        int        // bucket capacity (max burst)
}

/**
 * NewRateLimiterStore creates a store.
 *   - r: sustained request rate (e.g. 5 = 5 req/s)
 *   - b: burst allowance  (e.g. 10 = up to 10 consecutive requests)
*/
func NewRateLimiterStore(r rate.Limit, b int) *RateLimiterStore {
	store := &RateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		r:        r,
		b:        b,
	}

	/**
	 * Background goroutine: every minute, evict IPs not seen for 3 minutes.
	 * Without this, the map grows unboundedly as unique IPs hit the server.
	*/
	go store.cleanupLoop()

	return store
}

// getLimiter returns the limiter for the given IP, creating one if needed.
func (s *RateLimiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.limiters[ip]
	if !exists {
		limiter := rate.NewLimiter(s.r, s.b)
		s.limiters[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	entry.lastSeen = time.Now()
	return entry.limiter
}

// cleanupLoop runs every 60 seconds and evicts idle entries.
func (s *RateLimiterStore) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for ip, entry := range s.limiters {
			if time.Since(entry.lastSeen) > 3*time.Minute {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

/**
 * RateLimit returns a Gin middleware that enforces per-IP rate limiting.
 * 
 * Usage in router:
 * 
 * store := middleware.NewRateLimiterStore(5, 10) // 5 req/s, burst 10
 * router.Use(middleware.RateLimit(store))
*/
func RateLimit(store *RateLimiterStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := store.getLimiter(ip)

		/**
		 * Allow() is non-blocking — it either consumes a token or returns false.
		 * We use Allow() instead of Wait() to avoid holding the goroutine.
		*/
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "too many requests — please slow down",
			})
			return
		}

		c.Next()
	}
}

/**
 * StrictRateLimit is a tighter limiter for sensitive endpoints like /auth/login.
 * Creates its own internal store so it's independent of the global limiter.
 * 
 * r=1  → 1 request/second sustained
 * b=5  → allow a burst of 5 before throttling
*/
func StrictRateLimit() gin.HandlerFunc {
	store := NewRateLimiterStore(1, 5)
	return RateLimit(store)
}