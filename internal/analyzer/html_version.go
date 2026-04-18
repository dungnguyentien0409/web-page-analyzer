package analyzer

import (
	"strings"
)

func detectHTMLVersion(html []byte) string {
	htmlLower := strings.ToLower(string(html))

	switch {
	case strings.Contains(htmlLower, "<!doctype html>"):
		return "HTML5"

	case strings.Contains(htmlLower, "html 4.01"):
		return "HTML 4.01"

	case strings.Contains(htmlLower, "xhtml"):
		return "XHTML"

	default:
		return "Unknown"
	}
}
