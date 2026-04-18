package analyzer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func checkLoginForm(doc *goquery.Document) bool {
	hasLogin := false
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		if s.Find("input[type='password']").Length() > 0 {
			hasLogin = true
		}
		action, _ := s.Attr("action")
		if strings.Contains(strings.ToLower(action), "login") || strings.Contains(strings.ToLower(action), "signin") {
			hasLogin = true
		}
	})
	return hasLogin
}
