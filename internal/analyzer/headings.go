package analyzer

import (
	"github.com/PuerkitoBio/goquery"
)

func countHeadings(doc *goquery.Document) map[string]int {
	result := map[string]int{
		"h1": 0,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}
	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		result[goquery.NodeName(s)]++
	})
	return result
}
