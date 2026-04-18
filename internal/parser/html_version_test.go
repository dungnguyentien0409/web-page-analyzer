package parser

import "testing"

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{"HTML5", "<!DOCTYPE html><html>", "HTML5"},
		{"HTML 4.01", "public \"-//W3C//DTD HTML 4.01//EN\"", "HTML 4.01"},
		{"XHTML", "http://www.w3.org/1999/xhtml", "XHTML"},
		{"Unknown", "<html><body>", "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectHTMLVersion([]byte(tt.html))
			if got != tt.expected {
				t.Errorf("detectHTMLVersion() = %v, want %v", got, tt.expected)
			}
		})
	}
}
