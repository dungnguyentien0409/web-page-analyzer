package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiterBlocksRequests(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1) // 1 req/sec

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/analyze", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	// First request should pass
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)

	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}

	// Second request should be blocked
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}
}

func TestDifferentIPsHaveDifferentLimiters(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1)

	handler := limiter.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "127.0.0.1:1234"

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:1234"

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatal("different IP should not be limited")
	}
}