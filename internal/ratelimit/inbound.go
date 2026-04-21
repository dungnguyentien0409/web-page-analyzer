package ratelimit

import (
	"net"
	"net/http"
	"sync"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
	"golang.org/x/time/rate"
)

type InboundLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
	metrics  *metrics.Collector
}

type InboundConfig struct {
	RPS     float64
	Burst   int
	Metrics *metrics.Collector
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
		metrics:  cfg.Metrics,
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
			if i.metrics != nil {
				i.metrics.IncRateLimitRejection("inbound")
			}
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
