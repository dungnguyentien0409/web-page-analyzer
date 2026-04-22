package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/ratelimit"
)

func TestRateLimiterIntegration(t *testing.T) {
	t.Parallel()
	limiter := ratelimit.NewInboundLimiter(ratelimit.InboundConfig{
		RPS:   2,
		Burst: 2,
	})

	mux := http.NewServeMux()
	mux.Handle("/test", limiter.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)))

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{}

	blocked := false

	for i := 0; i < 10; i++ {
		resp, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("failed to send request: %v", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			_ = resp.Body.Close()
			blocked = true
			break
		}
		_ = resp.Body.Close()
	}

	if !blocked {
		t.Fatal("expected some requests to be rate limited")
	}
}

func TestOutboundRateLimiterIntegration(t *testing.T) {
	t.Parallel()
	// Create outbound limiter with very low limits for testing
	limiter := ratelimit.NewOutboundLimiter(ratelimit.OutboundConfig{
		GlobalRPS:   2,
		GlobalBurst: 2,
		HostRPS:     1,
		HostBurst:   1,
	})
	defer limiter.Stop()

	// Track request times
	var requestTimes []time.Time
	var mu sync.Mutex

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()

	// Test global rate limiting - burst should allow 2 immediate requests
	start := time.Now()
	for i := 0; i < 3; i++ {
		err := limiter.Wait(ctx, "test-host")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	elapsed := time.Since(start)

	// Third request should be rate limited (burst=2, rps=2)
	// So it should take at least ~500ms (1/rps = 500ms for 3rd request)
	if elapsed < 400*time.Millisecond {
		t.Errorf("expected rate limiting to delay requests, elapsed: %v", elapsed)
	}

	// Test per-host rate limiting
	requestTimes = nil

	// Make requests to different hosts concurrently
	done := make(chan bool, 2)
	go func() {
		for i := 0; i < 2; i++ {
			_ = limiter.Wait(ctx, "host-a")
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 2; i++ {
			_ = limiter.Wait(ctx, "host-b")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Test context cancellation
	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := limiter.Wait(ctx, "test-host")
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	})

	// Test empty host defaults to "unknown"
	t.Run("EmptyHost", func(t *testing.T) {
		err := limiter.Wait(ctx, "")
		if err != nil {
			t.Errorf("unexpected error for empty host: %v", err)
		}
	})
}
