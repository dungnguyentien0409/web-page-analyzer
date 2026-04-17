package parser

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
)

func ParseHTML(html []byte) (*goquery.Document, error) {
	reader := bytes.NewReader(html)

	doc, err := goquery.NewDocumentFromReader(reader)

	if err != nil {
		return nil, err
	}

	return doc, nil
}

func ExtractTitle(doc *goquery.Document) string {
	title := doc.Find("title").Text()
	return title
}