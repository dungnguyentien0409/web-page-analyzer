package analyzer

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCheckLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name:     "Has password input",
			html:     `<form><input type="password"></form>`,
			expected: true,
		},
		{
			name:     "Has login action",
			html:     `<form action="/auth/login"><input type="text"></form>`,
			expected: true,
		},
		{
			name:     "No login form",
			html:     `<form action="/search"><input type="text"></form>`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			got := checkLoginForm(doc)
			if got != tt.expected {
				t.Errorf("checkLoginForm() = %v, want %v", got, tt.expected)
			}
		})
	}
}
