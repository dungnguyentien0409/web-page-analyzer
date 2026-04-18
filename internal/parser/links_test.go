package parser

import (
	"context"
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
		<a href="/internal">Internal Duplicate</a>
		<a href="` + ts.URL + `/valid">External Valid</a>
		<a href="` + ts.URL + `/broken">External Broken</a>
		<a href="javascript:void(0)">JS</a>
		<a href="mailto:test@test.com">Mailto</a>
		<a href="#anchor">Anchor</a>
		<a>No Href</a>
		<a href="">Empty Href</a>
		<a href="%%">Malformed Href</a>
		<a href=" ://invalid-url ">Invalid URL</a>
		<input href="/not-an-a-tag">
		<a href="http://">Empty Host</a>
		<a href="` + ts.URL + `/valid">External Duplicate</a>
	`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	res, err := extractLinks(context.TODO(), doc, baseURL)
	if err != nil {
		t.Fatalf("extractLinks failed: %v", err)
	}
	if res.Internal != 1 {
		t.Errorf("expected 1 unique internal link, got %d", res.Internal)
	}
	if res.External != 2 {
		t.Errorf("expected 2 unique external links, got %d", res.External)
	}
	if res.Inaccessible != 2 {
		t.Errorf("expected 2 inaccessible links, got %d", res.Inaccessible)
	}
}
func TestIsLinkAccessible_Fail(t *testing.T) {
	if isLinkAccessible(context.TODO(), "http://non-existent-domain-12345.com") {
		t.Error("expected false for invalid domain")
	}
}
func TestIsLinkAccessible_MalformedURL(t *testing.T) {
	if isLinkAccessible(context.TODO(), ":") {
		t.Error("expected false for malformed URL")
	}
}
func TestIsLinkAccessible_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if isLinkAccessible(ctx, "http://example.com") {
		t.Error("expected false for canceled context")
	}
}
func TestExtractLinks_InvalidURL(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(""))
	_, err := extractLinks(context.TODO(), doc, " %%%% ")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
