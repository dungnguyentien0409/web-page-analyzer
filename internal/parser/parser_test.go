package parser

import "testing"

func TestParseHTML_Success(t *testing.T) {
	html := []byte("<html><head><title>Test</title></head><body></body></html>")

	doc, err := ParseHTML(html)

	if err != nil {
		t.Fatalf("expected no error, got %v",err)
	}

	if doc == nil {
		t.Fatal("expected document, got nil")
	}
}