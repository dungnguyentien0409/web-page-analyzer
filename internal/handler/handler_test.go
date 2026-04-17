package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

func setupTestTemplate() {
	testTemplate := template.Must(
		template.New("test").Parse(`
<html>
<body>
{{.Result}}
</body>
</html>
`),
	)
	SetTemplate(testTemplate)
}

func TestIndexHandler(t *testing.T) {
	setupTestTemplate()

	req := httptest.NewRequest(http.MethodGet,"/",nil)

	rr := httptest.NewRecorder()

	IndexHandler(rr,req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d",http.StatusOK,rr.Code)
	}
}

func TestAnalyzeHandler_MethodNotAllowed(t *testing.T) {
	setupTestTemplate()

	req := httptest.NewRequest(http.MethodGet,"/analyze",nil)

	rr := httptest.NewRecorder()

	AnalyzeHandler(rr,req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d",http.StatusMethodNotAllowed,rr.Code)
	}
}

func TestAnalyzeHandler_EmptyURL(t *testing.T) {
	setupTestTemplate()

	form := url.Values{}
	form.Add("url","")

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

	AnalyzeHandler(rr,req)

	if !strings.Contains(rr.Body.String(),"URL is required") {
		t.Errorf("expected URL error message")
	}
}

func TestAnalyzeHandler_FetchError(t *testing.T) {
	setupTestTemplate()

	fetchURL = func(url string) ([]byte,error) {
		return nil,fmt.Errorf("fetch failed")
	}

	defer func() {
		fetchURL = fetcher.FetchURL
	}()

	form := url.Values{}
	form.Add("url","https://example.com")

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

	AnalyzeHandler(rr,req)

	if !strings.Contains(rr.Body.String(),"fetch failed") {
		t.Errorf("expected fetch error message")
	}
}

func TestAnalyzeHandler_ParseError(t *testing.T) {
	setupTestTemplate()

	fetchURL = func(url string) ([]byte,error) {
		return []byte("<html>"),nil
	}

	parseHTML = func(html []byte) (*goquery.Document,error) {
		return nil,fmt.Errorf("parse failed")
	}

	defer func() {
		fetchURL = fetcher.FetchURL
		parseHTML = parser.ParseHTML
	}()

	form := url.Values{}
	form.Add("url","https://example.com")

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

	AnalyzeHandler(rr,req)

	if !strings.Contains(rr.Body.String(),"Failed to parse HTML") {
		t.Errorf("expected parse error message")
	}
}

func TestAnalyzeHandler_Success(t *testing.T) {
	setupTestTemplate()

	fetchURL = func(url string) ([]byte,error) {
		return []byte("<html><head><title>Test</title></head><body></body></html>"),nil
	}

	defer func() {
		fetchURL = fetcher.FetchURL
	}()

	form := url.Values{}
	form.Add("url","https://example.com")

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

	AnalyzeHandler(rr,req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d",http.StatusOK,rr.Code)
	}

	if !strings.Contains(rr.Body.String(),"Page title: Test") {
		t.Errorf("expected title in response")
	}
}