package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewInboundLimiter_Defaults(t *testing.T) {
	limiter := NewInboundLimiter(InboundConfig{})

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}

	if limiter.r <= 0 {
		t.Error("expected default RPS to be set")
	}

	if limiter.burst <= 0 {
		t.Error("expected default burst to be set")
	}
}

func TestNewInboundLimiter_CustomConfig(t *testing.T) {
	limiter := NewInboundLimiter(InboundConfig{
		RPS:   100,
		Burst: 50,
	})

	if limiter == nil {
		t.Fatal("expected limiter, got nil")
	}

	if limiter.r != 100 {
		t.Errorf("expected RPS 100, got %v", limiter.r)
	}

	if limiter.burst != 50 {
		t.Errorf("expected burst 50, got %d", limiter.burst)
	}
}

func TestInboundLimiter_getLimiter(t *testing.T) {
	limiter := NewInboundLimiter(InboundConfig{
		RPS:   10,
		Burst: 5,
	})

	// First call creates new limiter
	l1 := limiter.getLimiter("192.168.1.1")
	if l1 == nil {
		t.Fatal("expected limiter, got nil")
	}

	// Second call returns same limiter
	l2 := limiter.getLimiter("192.168.1.1")
	if l1 != l2 {
		t.Error("expected same limiter for same IP")
	}

	// Different IP gets different limiter
	l3 := limiter.getLimiter("10.0.0.1")
	if l1 == l3 {
		t.Error("expected different limiter for different IP")
	}

	// Verify map has both entries
	limiter.mu.Lock()
	count := len(limiter.limiters)
	limiter.mu.Unlock()

	if count != 2 {
		t.Errorf("expected 2 limiters, got %d", count)
	}
}

func TestInboundLimiter_Middleware_Allows(t *testing.T) {
	limiter := NewInboundLimiter(InboundConfig{
		RPS:   100,
		Burst: 10,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	middleware := limiter.Middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestInboundLimiter_Middleware_Blocks(t *testing.T) {
	// Very low limit
	limiter := NewInboundLimiter(InboundConfig{
		RPS:   1,
		Burst: 1,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := limiter.Middleware(handler)

	// First request should pass
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected first request to pass with 200, got %d", rec.Code)
	}

	// Multiple rapid requests should be blocked
	blocked := false
	for i := 0; i < 10; i++ {
		rec = httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			blocked = true
			break
		}
	}

	if !blocked {
		t.Error("expected some requests to be rate limited")
	}
}

func TestInboundLimiter_Middleware_MultipleIPs(t *testing.T) {
	limiter := NewInboundLimiter(InboundConfig{
		RPS:   1,
		Burst: 1,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := limiter.Middleware(handler)

	// First IP
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	middleware.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("expected IP1 first request to pass, got %d", rec1.Code)
	}

	// Different IP should have separate limit
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	rec2 := httptest.NewRecorder()
	middleware.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("expected IP2 first request to pass, got %d", rec2.Code)
	}
}