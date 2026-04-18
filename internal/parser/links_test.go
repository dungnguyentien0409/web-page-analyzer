package parser

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractLinks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/broken" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	baseURL := "https://example.com"
	html := `
		<a href="/internal">Internal</a>
		<a href="` + ts.URL + `/valid">External Valid</a>
		<a href="` + ts.URL + `/broken">External Broken</a>
		<a href="javascript:void(0)">JS</a>
		<a href="mailto:test@test.com">Mailto</a>
		<a href="#anchor">Anchor</a>
		<a>No Href</a>
		<a href=" ://invalid-url ">Invalid URL</a>
		<input href="/not-an-a-tag">
	`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	res, err := extractLinks(doc, baseURL)
	if err != nil {
		t.Fatalf("extractLinks failed: %v", err)
	}
	if res.Internal != 1 {
		t.Errorf("expected 1 internal link, got %d", res.Internal)
	}
	if res.External != 2 {
		t.Errorf("expected 2 external links, got %d", res.External)
	}
	if res.Inaccessible != 2 {
		t.Errorf("expected 2 inaccessible links, got %d", res.Inaccessible)
	}
}
func TestIsLinkAccessible_Fail(t *testing.T) {
	if isLinkAccessible("http://non-existent-domain-12345.com") {
		t.Error("expected false for invalid domain")
	}
}
func TestExtractLinks_InvalidURL(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(""))
	_, err := extractLinks(doc, " %%%% ")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
