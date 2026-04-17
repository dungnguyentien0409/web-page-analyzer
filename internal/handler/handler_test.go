package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAnalyzeHandler_MethodNotAllowed(t *testing.T) {

	req := httptest.NewRequest(
		http.MethodGet,
		"/analyze",
		nil,
	)

	rr := httptest.NewRecorder()

	AnalyzeHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {

		t.Errorf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rr.Code,
		)

	}

}

func TestAnalyzeHandler_EmptyURL(t *testing.T) {

	form := url.Values{}
	form.Add("url", "")

	req := httptest.NewRequest(
		http.MethodPost,
		"/analyze",
		strings.NewReader(form.Encode()),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rr := httptest.NewRecorder()

	AnalyzeHandler(rr, req)

	if rr.Code != http.StatusBadRequest {

		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rr.Code,
		)

	}

}

func TestAnalyzeHandler_Success(t *testing.T) {

	form := url.Values{}
	form.Add("url", "https://example.com")

	req := httptest.NewRequest(
		http.MethodPost,
		"/analyze",
		strings.NewReader(form.Encode()),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rr := httptest.NewRecorder()

	AnalyzeHandler(rr, req)

	if rr.Code != http.StatusOK {

		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			rr.Code,
		)

	}

	expected := "Analyze endpoint working"

	if strings.TrimSpace(rr.Body.String()) != expected {

		t.Errorf(
			"expected body %q, got %q",
			expected,
			rr.Body.String(),
		)

	}

}