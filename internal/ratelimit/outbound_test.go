package ratelimit

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestNewOutboundLimiter_Defaults(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{})
	defer limiter.Stop()

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}

	if limiter.global == nil {
		t.Error("expected global limiter to be set")
	}

	if limiter.hostRPS <= 0 {
		t.Error("expected default hostRPS to be set")
	}

	if limiter.hostBurst <= 0 {
		t.Error("expected default hostBurst to be set")
	}
}

func TestNewOutboundLimiter_CustomConfig(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	limiter := NewOutboundLimiter(OutboundConfig{
		GlobalRPS:   500,
		GlobalBurst: 100,
		HostRPS:     50,
		HostBurst:   10,
		HostTTL:     5 * time.Minute,
		Logger:      logger,
	})
	defer limiter.Stop()

	if limiter.hostRPS != 50 {
		t.Errorf("expected hostRPS 50, got %d", limiter.hostRPS)
	}

	if limiter.hostBurst != 10 {
		t.Errorf("expected hostBurst 10, got %d", limiter.hostBurst)
	}

	if limiter.ttl != 5*time.Minute {
		t.Errorf("expected TTL 5m, got %v", limiter.ttl)
	}
}

func TestOutboundLimiter_Wait_Success(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{
		GlobalRPS:   100,
		GlobalBurst: 10,
		HostRPS:     10,
		HostBurst:   5,
	})
	defer limiter.Stop()

	ctx := context.Background()

	err := limiter.Wait(ctx, "example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOutboundLimiter_Wait_EmptyHost(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{
		GlobalRPS:   100,
		GlobalBurst: 10,
		HostRPS:     10,
		HostBurst:   5,
	})
	defer limiter.Stop()

	ctx := context.Background()

	// Empty host should use "unknown"
	err := limiter.Wait(ctx, "")
	if err != nil {
		t.Errorf("expected no error with empty host, got %v", err)
	}
}

func TestOutboundLimiter_Wait_ContextCancelled(t *testing.T) {
	// Very low rate limit that will block
	limiter := NewOutboundLimiter(OutboundConfig{
		GlobalRPS:   1,
		GlobalBurst: 1,
		HostRPS:     1,
		HostBurst:   1,
	})
	defer limiter.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	// Use up the tokens
	_ = limiter.Wait(ctx, "test.com")

	// Cancel context
	cancel()

	// This should fail because context is cancelled
	err := limiter.Wait(ctx, "test.com")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestOutboundLimiter_getHostLimiter_SameHost(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{
		HostRPS:   10,
		HostBurst: 5,
	})
	defer limiter.Stop()

	l1 := limiter.getHostLimiter("example.com")
	l2 := limiter.getHostLimiter("example.com")

	if l1 != l2 {
		t.Error("expected same limiter for same host")
	}
}

func TestOutboundLimiter_getHostLimiter_DifferentHosts(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{
		HostRPS:   10,
		HostBurst: 5,
	})
	defer limiter.Stop()

	l1 := limiter.getHostLimiter("example.com")
	l2 := limiter.getHostLimiter("google.com")

	if l1 == l2 {
		t.Error("expected different limiters for different hosts")
	}
}

func TestOutboundLimiter_doCleanup(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{
		HostRPS:   10,
		HostBurst: 5,
		HostTTL:   100 * time.Millisecond,
	})
	defer limiter.Stop()

	// Add some hosts
	limiter.getHostLimiter("example.com")
	limiter.getHostLimiter("google.com")

	// Wait for TTL to pass
	time.Sleep(150 * time.Millisecond)

	// Run cleanup
	limiter.doCleanup()

	// Check that hosts were cleaned
	count := 0
	limiter.hosts.Range(func(_, _ interface{}) bool {
		count++
		return true
	})

	if count != 0 {
		t.Errorf("expected 0 hosts after cleanup, got %d", count)
	}
}

func TestOutboundLimiter_doCleanup_KeepsActive(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{
		HostRPS:   10,
		HostBurst: 5,
		HostTTL:   1 * time.Second,
	})
	defer limiter.Stop()

	// Add host
	limiter.getHostLimiter("example.com")

	// Wait a bit but not enough for TTL
	time.Sleep(50 * time.Millisecond)

	// Run cleanup
	limiter.doCleanup()

	// Check that host is still there
	count := 0
	limiter.hosts.Range(func(_, _ interface{}) bool {
		count++
		return true
	})

	if count != 1 {
		t.Errorf("expected 1 host to remain, got %d", count)
	}
}

func TestOutboundLimiter_Concurrent(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{
		GlobalRPS:   1000,
		GlobalBurst: 100,
		HostRPS:     100,
		HostBurst:   10,
	})
	defer limiter.Stop()

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			host := "example.com"
			if i%2 == 0 {
				host = "google.com"
			}

			if err := limiter.Wait(context.Background(), host); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutboundLimiter_Stop(t *testing.T) {
	limiter := NewOutboundLimiter(OutboundConfig{})

	// Stop should not block indefinitely
	done := make(chan struct{})
	go func() {
		limiter.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Stop blocked for too long")
	}
}
