package handler

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/analyzer"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/middleware"
)

type mockFetcher struct {
	fn func(context.Context, string) ([]byte, error)
}

func (m *mockFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	return m.fn(ctx, url)
}

type errorResponseWriter struct {
	httptest.ResponseRecorder
}

func (e *errorResponseWriter) Write(b []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}
func setupTestHandler() *Handler {
	t := template.Must(
		template.New("test").Parse(`
<html>
<body>{{if .Result}}Page title: {{.Result.Title}}{{end}}{{if .Error}}{{.Error}}{{end}}
</body>
</html>
`),
	)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mc := metrics.NewCollector()
	return NewHandler(HandlerConfig{
		Template: t,
		Fetcher:  fetcher.NewDefaultFetcher(fetcher.FetcherConfig{}, logger, mc),
		Analyzer: analyzer.NewDefaultAnalyzer(analyzer.AnalyzerConfig{
			Logger:      logger,
			RetryCount:  3,
			WorkerCount: 20,
			Metrics:     mc,
		}),
		RequestTimeout: 5 * time.Second,
		Logger:         logger,
		Metrics:        mc,
	})
}
func TestIndexHandler(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.IndexHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
func TestAnalyzeHandler_MethodNotAllowed(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analyze", nil)
	rr := httptest.NewRecorder()
	h.AnalyzeHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
func TestAnalyzeHandler_EmptyURL(t *testing.T) {
	h := setupTestHandler()
	form := url.Values{}
	form.Add("url", "")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.AnalyzeHandler(rr, req)
	if !strings.Contains(rr.Body.String(), "URL is required") {
		t.Errorf("expected URL error message")
	}
}
func TestAnalyzeHandler_FetchError(t *testing.T) {
	h := setupTestHandler()
	h.fetcher = &mockFetcher{fn: func(ctx context.Context, url string) ([]byte, error) {
		return nil, fmt.Errorf("fetch failed")
	}}
	form := url.Values{}
	form.Add("url", "https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.AnalyzeHandler(rr, req)
	if !strings.Contains(rr.Body.String(), "fetch failed") {
		t.Errorf("expected fetch error message")
	}
}
func TestAnalyzeHandler_ParseError(t *testing.T) {
	h := setupTestHandler()
	h.fetcher = &mockFetcher{fn: func(ctx context.Context, url string) ([]byte, error) {
		return []byte("<html>"), nil
	}}
	h.analyzePage = func(ctx context.Context, req analyzer.AnalysisRequest) (*analyzer.AnalysisResult, error) {
		return nil, fmt.Errorf("analysis failed")
	}
	form := url.Values{}
	form.Add("url", "https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.AnalyzeHandler(rr, req)
	if !strings.Contains(rr.Body.String(), "Failed to parse HTML") {
		t.Errorf("expected parse error message")
	}
}
func TestAnalyzeHandler_Success(t *testing.T) {
	h := setupTestHandler()
	h.fetcher = &mockFetcher{fn: func(ctx context.Context, url string) ([]byte, error) {
		return []byte("<html><head><title>Test</title></head><body></body></html>"), nil
	}}
	form := url.Values{}
	form.Add("url", "https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// Test with request ID middleware
	handler := middleware.RequestID(http.HandlerFunc(h.AnalyzeHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Page title: Test") {
		t.Errorf("expected title in response")
	}
	// Check that request ID was set
	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}
func TestAnalyzeHandler_TemplateError(t *testing.T) {
	h := setupTestHandler()
	h.tmpl = template.Must(template.New("error").Option("missingkey=error").Parse("{{.InvalidField}}"))
	h.fetcher = &mockFetcher{fn: func(ctx context.Context, url string) ([]byte, error) {
		return []byte("<html></html>"), nil
	}}
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader("url=https://example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.AnalyzeHandler(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}
func TestAnalyzeHandler_WriteError(t *testing.T) {
	h := setupTestHandler()
	h.fetcher = &mockFetcher{fn: func(ctx context.Context, url string) ([]byte, error) {
		return []byte("<html></html>"), nil
	}}
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader("url=https://example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ew := &errorResponseWriter{*httptest.NewRecorder()}
	h.AnalyzeHandler(ew, req)
}
