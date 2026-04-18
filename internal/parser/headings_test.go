package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCountHeadings(t *testing.T) {
	html := `<h1>Title</h1><h2>Sub1</h2><h2>Sub2</h2><h3>Deep</h3>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	counts := countHeadings(doc)
	if counts["h1"] != 1 {
		t.Errorf("expected 1 h1, got %d", counts["h1"])
	}
	if counts["h2"] != 2 {
		t.Errorf("expected 2 h2, got %d", counts["h2"])
	}
	if counts["h3"] != 1 {
		t.Errorf("expected 1 h3, got %d", counts["h3"])
	}
	if counts["h6"] != 0 {
		t.Errorf("expected 0 h6, got %d", counts["h6"])
	}
}
