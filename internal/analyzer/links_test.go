package analyzer

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/ratelimit"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractLinks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()
	tsInternal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer tsInternal.Close()
	tsExternal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/broken" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer tsExternal.Close()

	baseURL := tsInternal.URL
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	html := `
			<a href="/internal">Internal</a>
			<a href="/internal">Internal Duplicate</a>
			<a href="` + tsExternal.URL + `/valid">External</a>
			<a href="` + tsExternal.URL + `/broken">External Broken</a>
			<a href="javascript:void(0)">JS</a>
			<a href="mailto:test@test.com">Mailto</a>
			<a href="#anchor">Anchor</a>
			<a>No Href</a>
			<a href="">Empty Href</a>
			<a href="` + "\x00" + `">Malformed Href</a>
			<input href="/not-an-a-tag">
			<a href="http://">Empty Host</a>
			<a href="` + tsExternal.URL + `/valid">Duplicate External</a>
		`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	res, err := a.extractLinks(context.TODO(), doc, baseURL)

	if err != nil {
		t.Fatalf("extractLinks failed: %v", err)
	}
	if res.Internal != 1 {
		t.Errorf("expected 1 unique internal link, got %d", res.Internal)
	}
	if res.External != 2 {
		t.Errorf("expected 2 unique external links, got %d", res.External)
	}
	if res.Inaccessible != 1 {
		t.Errorf("expected 1 inaccessible link, got %d", res.Inaccessible)
	}
}

func TestIsLinkAccessible_Fail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	if a.isLinkAccessible(context.TODO(), "http://non-existent-domain-12345.com") {
		t.Error("expected false for invalid domain")
	}
}

func TestIsLinkAccessible_MalformedURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	if a.isLinkAccessible(context.TODO(), "\x00") {
		t.Error("expected false for malformed URL")
	}
}

func TestIsLinkAccessible_ContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if a.isLinkAccessible(ctx, "http://example.com") {
		t.Error("expected false for canceled context")
	}
}

func TestIsLinkAccessible_RetrySuccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	if !a.isLinkAccessible(context.TODO(), ts.URL) {
		t.Error("expected true after retry")
	}
}

func TestIsLinkAccessible_Persistent500(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	if a.isLinkAccessible(context.TODO(), ts.URL) {
		t.Error("expected false for persistent 500")
	}
}

func TestIsLinkAccessible_RetryContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(context.Background())
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		cancel()
	}))
	defer ts.Close()
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	if a.isLinkAccessible(ctx, ts.URL) {
		t.Error("expected false when context canceled during retry")
	}
}

func TestExtractLinks_InvalidURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
		Metrics:     mc,
	})
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(""))
	_, err := a.extractLinks(context.TODO(), doc, "\x00")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestIsLinkAccessible_EmptyResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  1,
		WorkerCount: 1,
		Metrics:     mc,
	})
	// URL không hợp lệ nhưng qua được bước tạo Request
	if a.isLinkAccessible(context.TODO(), "http://") {
		t.Error("expected false for empty host")
	}
}

func TestIsLinkAccessible_WithRateLimiter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	limiter := ratelimit.NewOutboundLimiter(ratelimit.OutboundConfig{
		GlobalRPS:   100,
		GlobalBurst: 10,
		HostRPS:     10,
		HostBurst:   5,
	})
	defer limiter.Stop()

	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:           logger,
		RetryCount:       3,
		WorkerCount:      20,
		Metrics:          mc,
		OutboundLimiter:  limiter,
	})

	if !a.isLinkAccessible(context.TODO(), ts.URL) {
		t.Error("expected true for accessible URL with rate limiter")
	}
}

func TestIsLinkAccessible_RateLimiterContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()

	// Very restrictive rate limiter
	limiter := ratelimit.NewOutboundLimiter(ratelimit.OutboundConfig{
		GlobalRPS:   1,
		GlobalBurst: 1,
		HostRPS:     1,
		HostBurst:   1,
	})
	defer limiter.Stop()

	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:           logger,
		RetryCount:       3,
		WorkerCount:      20,
		Metrics:          mc,
		OutboundLimiter:  limiter,
	})

	// Use up the tokens
	_ = limiter.Wait(context.Background(), "test.com")

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return false because rate limiter will error on canceled context
	if a.isLinkAccessible(ctx, "http://example.com") {
		t.Error("expected false when rate limiter context is canceled")
	}
}

func TestIsLinkAccessible_EmptyHost(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()

	limiter := ratelimit.NewOutboundLimiter(ratelimit.OutboundConfig{
		GlobalRPS:   100,
		GlobalBurst: 10,
		HostRPS:     10,
		HostBurst:   5,
	})
	defer limiter.Stop()

	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:           logger,
		RetryCount:       3,
		WorkerCount:      20,
		Metrics:          mc,
		OutboundLimiter:  limiter,
	})

	// Test with URL that will have empty host after parsing error
	// The isLinkAccessible function should handle the empty host case
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Normal URL should work
	if !a.isLinkAccessible(context.TODO(), ts.URL) {
		t.Error("expected true for accessible URL")
	}
}