package ratelimit

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type OutboundLimiter struct {
	global *rate.Limiter
	hosts  sync.Map // map[string]*hostEntry

	hostRPS   int
	hostBurst int

	// cleanup
	stopCh chan struct{}
	waitCh chan struct{}
	ttl    time.Duration
	logger *slog.Logger
}

type hostEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type OutboundConfig struct {
	GlobalRPS   int
	GlobalBurst int
	HostRPS     int
	HostBurst   int
	HostTTL     time.Duration
	Logger      *slog.Logger
}

func NewOutboundLimiter(cfg OutboundConfig) *OutboundLimiter {

	// Defaults

	if cfg.GlobalRPS <= 0 {
		cfg.GlobalRPS = 1000
	}

	if cfg.GlobalBurst <= 0 {
		cfg.GlobalBurst = 100
	}

	if cfg.HostRPS <= 0 {
		cfg.HostRPS = 100
	}

	if cfg.HostBurst <= 0 {
		cfg.HostBurst = 20
	}

	if cfg.HostTTL <= 0 {
		cfg.HostTTL = 10 * time.Minute
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	o := &OutboundLimiter{
		global: rate.NewLimiter(
			rate.Limit(cfg.GlobalRPS),
			cfg.GlobalBurst,
		),
		hostRPS:   cfg.HostRPS,
		hostBurst: cfg.HostBurst,
		stopCh:    make(chan struct{}),
		waitCh:    make(chan struct{}),
		ttl:       cfg.HostTTL,
		logger:    cfg.Logger,
	}

	go o.cleanup()

	return o
}

func (o *OutboundLimiter) Wait(ctx context.Context, host string) error {

	if host == "" {
		host = "unknown"
	}

	// Global rate limit
	if err := o.global.Wait(ctx); err != nil {
		return err
	}

	// Per-host rate limit
	limiter := o.getHostLimiter(host)

	return limiter.Wait(ctx)
}

func (o *OutboundLimiter) getHostLimiter(host string) *rate.Limiter {

	// Check existing
	if v, ok := o.hosts.Load(host); ok {
		entry := v.(*hostEntry)
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	// Create new entry
	entry := &hostEntry{
		limiter:  rate.NewLimiter(rate.Limit(o.hostRPS), o.hostBurst),
		lastSeen: time.Now(),
	}

	// Store safely
	actual, _ := o.hosts.LoadOrStore(host, entry)

	return actual.(*hostEntry).limiter
}

func (o *OutboundLimiter) cleanup() {

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer close(o.waitCh)

	for {
		select {
		case <-ticker.C:
			o.doCleanup()
		case <-o.stopCh:
			return
		}
	}
}

func (o *OutboundLimiter) doCleanup() {

	now := time.Now()
	cleaned := 0

	o.hosts.Range(func(key, value interface{}) bool {
		entry := value.(*hostEntry)

		if now.Sub(entry.lastSeen) > o.ttl {
			o.hosts.Delete(key)
			cleaned++
		}

		return true
	})

	if cleaned > 0 {
		o.logger.Debug("outbound limiter cleanup", "cleaned", cleaned)
	}
}

func (o *OutboundLimiter) Stop() {
	close(o.stopCh)
	<-o.waitCh
}