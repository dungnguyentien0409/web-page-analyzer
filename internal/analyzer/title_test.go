package analyzer

import (
	"io"
	"log/slog"
	"testing"
)

func TestExtractTitle_WithTitle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
	})
	html := []byte("<html><head><title>Hello</title></head></html>")
	doc, err := a.ParseHTML(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title := ExtractTitle(doc)
	if title != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", title)
	}
}
func TestExtractTitle_NoTitle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
	})
	html := []byte("<html><head></head></html>")
	doc, err := a.ParseHTML(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title := ExtractTitle(doc)
	if title != "" {
		t.Errorf("expected empty title, got '%s'", title)
	}
}
func TestExtractTitle_MultipleTitles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
	})
	html := []byte("<html><head><title>First</title><title>Second</title></head></html>")
	doc, err := a.ParseHTML(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title := ExtractTitle(doc)
	if title != "FirstSecond" {
		t.Errorf("expected 'FirstSecond', got '%s'", title)
	}
}
