package fetcher

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
)

func TestFetchURL_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test response"))
		}),
	)
	defer server.Close()
	mc := metrics.NewCollector()
	f := NewDefaultFetcher(slog.New(slog.NewTextHandler(io.Discard, nil)), mc)
	body, err := f.Fetch(context.TODO(), server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := "test response"
	if string(body) != expected {
		t.Errorf("expected body %q, got %q", expected, string(body))
	}
}
func TestFetchURL_HTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	defer server.Close()
	mc := metrics.NewCollector()
	f := NewDefaultFetcher(slog.New(slog.NewTextHandler(io.Discard, nil)), mc)
	_, err := f.Fetch(context.TODO(), server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
func TestFetchURL_InvalidURL(t *testing.T) {
	mc := metrics.NewCollector()
	f := NewDefaultFetcher(slog.New(slog.NewTextHandler(io.Discard, nil)), mc)
	_, err := f.Fetch(context.TODO(), "://bad-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
func TestFetchURL_ContextCanceled(t *testing.T) {
	mc := metrics.NewCollector()
	f := NewDefaultFetcher(slog.New(slog.NewTextHandler(io.Discard, nil)), mc)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := f.Fetch(ctx, "http://example.com")
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}
func TestFetchURL_ReadError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("hijacking not supported")
			}
			conn, _, err := hijacker.Hijack()
			if err != nil {
				t.Fatal(err)
			}
			if err := conn.Close(); err != nil {
				t.Logf("error closing connection: %v", err)
			}
		}),
	)
	defer server.Close()
	mc := metrics.NewCollector()
	f := NewDefaultFetcher(slog.New(slog.NewTextHandler(io.Discard, nil)), mc)
	_, err := f.Fetch(context.TODO(), server.URL)
	if err == nil {
		t.Fatal("expected read error")
	}
}
