package analyzer

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

func countHeadings(doc *goquery.Document) map[string]int {
	result := make(map[string]int)

	for i := 1; i <= 6; i++ {
		tag := fmt.Sprintf("h%d", i)

		result[tag] = doc.Find(tag).Length()
	}

	return result
}
