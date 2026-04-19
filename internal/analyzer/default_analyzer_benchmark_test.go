package analyzer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkAnalyzePage(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  1,
		WorkerCount: 20,
	})
	html := "<html><head><title>Bench</title></head><body>"
	for i := 0; i < 50; i++ {
		html += fmt.Sprintf("<h1>Heading %d</h1>", i)
		html += fmt.Sprintf("<a href=\"%s/%d\">Link %d</a>", ts.URL, i, i)
	}
	html += "</body></html>"
	req := PageAnalysisRequest{
		HTML: []byte(html),
		URL:  ts.URL,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.AnalyzePage(context.Background(), req)
	}
}
