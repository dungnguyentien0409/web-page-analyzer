package analyzer

import (
	"github.com/PuerkitoBio/goquery"
)

func countHeadings(doc *goquery.Document) map[string]int {
	result := make(map[string]int)
	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		result[goquery.NodeName(s)]++
	})
	return result
}
