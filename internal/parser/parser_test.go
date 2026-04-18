package parser

import (
	"errors"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseHTML_Success(t *testing.T) {
	a := NewAnalyzer()
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
	a := &Analyzer{
		documentProvider: func(r io.Reader) (*goquery.Document, error) {
			return nil, errors.New("parse error")
		},
	}
	_, err := a.ParseHTML([]byte("<html>"))
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestExtractTitle_WithTitle(t *testing.T) {
	a := NewAnalyzer()
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
	a := NewAnalyzer()
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
	a := NewAnalyzer()
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
func TestAnalyzePage(t *testing.T) {
	a := NewAnalyzer()
	html := []byte(`<html><head><title>Test</title></head><body><form><input type="password"></form></body></html>`)
	url := "https://test.com"
	result, err := a.AnalyzePage(PageSource{HTML: html, URL: url})
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
	a := NewAnalyzer()
	html := []byte(`<html></html>`)
	_, err := a.AnalyzePage(PageSource{HTML: html, URL: " %%%% "})
	if err == nil {
		t.Error("expected error for invalid base URL")
	}
}
func TestAnalyzePage_ParseError(t *testing.T) {
	a := &Analyzer{
		documentProvider: func(r io.Reader) (*goquery.Document, error) {
			return nil, errors.New("parse error")
		},
	}
	_, err := a.AnalyzePage(PageSource{HTML: []byte("<html>"), URL: "https://test.com"})
	if err == nil {
		t.Error("expected error from ParseHTML")
	}
}
