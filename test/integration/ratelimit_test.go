package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/middleware"
	"golang.org/x/time/rate"
)

func TestRateLimiterIntegration(t *testing.T) {
	limiter := middleware.NewIPRateLimiter(rate.Limit(2), 2)

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
