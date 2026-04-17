package parser

import (
	"errors"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseHTML_Success(t *testing.T) {
	html := []byte("<html><head><title>Test</title></head></html>")

	doc, err := ParseHTML(html)

	if err != nil {
		t.Fatalf("expected no error, got %v",err)
	}

	if doc == nil {
		t.Fatal("expected document, got nil")
	}
}

func TestParseHTML_Error(t *testing.T) {
	newDocumentFromReader = func(r io.Reader) (*goquery.Document,error) {
		return nil,errors.New("parse error")
	}

	defer func() {
		newDocumentFromReader = goquery.NewDocumentFromReader
	}()

	_, err := ParseHTML([]byte("<html>"))

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractTitle_WithTitle(t *testing.T) {
	html := []byte("<html><head><title>Hello</title></head></html>")

	doc, err := ParseHTML(html)

	if err != nil {
		t.Fatalf("unexpected error: %v",err)
	}

	title := ExtractTitle(doc)

	if title != "Hello" {
		t.Errorf("expected 'Hello', got '%s'",title)
	}
}

func TestExtractTitle_NoTitle(t *testing.T) {
	html := []byte("<html><head></head></html>")

	doc, err := ParseHTML(html)

	if err != nil {
		t.Fatalf("unexpected error: %v",err)
	}

	title := ExtractTitle(doc)

	if title != "" {
		t.Errorf("expected empty title, got '%s'",title)
	}
}

func TestExtractTitle_MultipleTitles(t *testing.T) {
	html := []byte("<html><head><title>First</title><title>Second</title></head></html>")

	doc, err := ParseHTML(html)

	if err != nil {
		t.Fatalf("unexpected error: %v",err)
	}

	title := ExtractTitle(doc)

	if title != "FirstSecond" {
		t.Errorf("expected 'FirstSecond', got '%s'",title)
	}
}