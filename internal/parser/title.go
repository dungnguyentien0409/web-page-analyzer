package parser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ExtractTitle(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find("title").Text())
}
