package analyzer

import (
	"bytes"
)

func detectHTMLVersion(html []byte) string {
	htmlLower := bytes.ToLower(html)
	switch {
	case bytes.Contains(htmlLower, []byte("<!doctype html>")):
		return "HTML5"
	case bytes.Contains(htmlLower, []byte("html 4.01")):
		return "HTML 4.01"
	case bytes.Contains(htmlLower, []byte("xhtml")):
		return "XHTML"
	default:
		return "Unknown"
	}
}
