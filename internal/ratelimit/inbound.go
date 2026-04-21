package ratelimit

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type InboundLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

type InboundConfig struct {
	RPS   float64
	Burst int
}

func NewInboundLimiter(cfg InboundConfig) *InboundLimiter {

	// Defaults
	if cfg.RPS <= 0 {
		cfg.RPS = 10
	}

	if cfg.Burst <= 0 {
		cfg.Burst = 20
	}

	return &InboundLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(cfg.RPS),
		burst:    cfg.Burst,
	}
}

func (i *InboundLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

func (i *InboundLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		limiter := i.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
