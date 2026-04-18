package analyzer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseHTML_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
	})
	html := []byte("<html><head><title>Test</title></head></html>")
	doc, err := a.ParseHTML(html)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if doc == nil {
		t.Fatal("expected document, got nil")
	}
}
func TestParseHTML_Error(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := &DefaultAnalyzer{
		logger: logger,
		documentProvider: func(r io.Reader) (*goquery.Document, error) {
			return nil, errors.New("parse error")
		},
	}
	_, err := a.ParseHTML([]byte("<html>"))
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestAnalyzePage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
	})
	html := []byte(`<html><head><title>Test</title></head><body><form><input type="password"></form></body></html>`)
	url := "https://test.com"
	result, err := a.AnalyzePage(context.TODO(), PageAnalysisRequest{HTML: html, URL: url})
	if err != nil {
		t.Fatalf("AnalyzePage failed: %v", err)
	}
	if result.Title != "Test" {
		t.Errorf("expected Title 'Test', got '%s'", result.Title)
	}
	if !result.ContainsLoginForm {
		t.Error("expected ContainsLoginForm to be true")
	}
}
func TestAnalyzePage_InvalidURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := NewDefaultAnalyzer(AnalyzerConfig{
		Logger:      logger,
		RetryCount:  3,
		WorkerCount: 20,
	})
	html := []byte(`<html></html>`)
	_, err := a.AnalyzePage(context.TODO(), PageAnalysisRequest{HTML: html, URL: " %%%% "})
	if err == nil {
		t.Error("expected error for invalid base URL")
	}
}
func TestAnalyzePage_ParseError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := &DefaultAnalyzer{
		logger: logger,
		documentProvider: func(r io.Reader) (*goquery.Document, error) {
			return nil, errors.New("parse error")
		},
	}
	_, err := a.AnalyzePage(context.TODO(), PageAnalysisRequest{HTML: []byte("<html>"), URL: "https://test.com"})
	if err == nil {
		t.Error("expected error from ParseHTML")
	}
}
