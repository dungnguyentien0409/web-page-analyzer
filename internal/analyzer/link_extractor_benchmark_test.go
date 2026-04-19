package analyzer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func BenchmarkExtractLinks(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  1,
		WorkerCount: 50,
	})
	var sb strings.Builder
	sb.WriteString("<html><body>")
	for i := 0; i < 200; i++ {
		sb.WriteString(fmt.Sprintf("<a href=\"%s/%d\">External Link %d</a>", ts.URL, i, i))
		sb.WriteString(fmt.Sprintf("<a href=\"/internal/%d\">Internal Link %d</a>", i, i))
	}
	sb.WriteString("</body></html>")
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(sb.String()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.extractLinks(context.Background(), doc, ts.URL)
	}
}
