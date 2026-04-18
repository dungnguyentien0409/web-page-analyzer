package integration

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/handler"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

func TestAnalyze_Integration(t *testing.T) {
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/broken" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer externalServer.Close()
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html, err := os.ReadFile("../testdata/sample_page.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		content := string(html)
		content = strings.ReplaceAll(content, "http://external-valid.com", externalServer.URL+"/valid")
		content = strings.ReplaceAll(content, "http://external-broken", externalServer.URL+"/broken")
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(content))
	}))
	defer targetServer.Close()
	tmpl, err := template.ParseFiles("../../web/templates/index.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	f := fetcher.NewDefaultFetcher(logger)
	p := parser.NewDefaultAnalyzer(logger)
	h := handler.NewHandler(handler.HandlerConfig{
		Template:       tmpl,
		Fetcher:        f,
		Analyzer:       p,
		RequestTimeout: 10 * time.Second,
		Logger:         logger,
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.IndexHandler)
	mux.HandleFunc("/analyze", h.AnalyzeHandler)
	appServer := httptest.NewServer(mux)
	defer appServer.Close()
	t.Run("Successful Analysis", func(t *testing.T) {
		form := url.Values{}
		form.Add("url", targetServer.URL)
		resp, err := http.PostForm(appServer.URL+"/analyze", form)
		if err != nil {
			t.Fatalf("failed to send request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status OK, got %v", resp.Status)
		}
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Full Integration Test") {
			t.Error("response missing page title")
		}
		if !strings.Contains(bodyStr, "HTML5") {
			t.Error("response missing HTML version")
		}
		if !strings.Contains(bodyStr, "Internal Links:</td><td class=\"value-col\">1</td>") {
			t.Error("internal links count mismatch")
		}
		if !strings.Contains(bodyStr, "External Links:</td><td class=\"value-col\">2</td>") {
			t.Error("external links count mismatch")
		}
		if !strings.Contains(bodyStr, "Inaccessible Links:</td><td class=\"value-col\" style=\"color: #b91c1c\">1</td>") {
			t.Error("inaccessible links count mismatch")
		}
		for i := 1; i <= 6; i++ {
			headingTag := "h" + string(rune(48+i))
			if !strings.Contains(bodyStr, "<td class=\"label-col\">"+headingTag+"</td><td class=\"value-col\">1</td>") {
				t.Errorf("missing or incorrect count for heading %s", headingTag)
			}
		}
		if !strings.Contains(bodyStr, "Yes") {
			t.Error("response missing login form indicator")
		}
	})
	t.Run("Failed Analysis - Unreachable URL", func(t *testing.T) {
		form := url.Values{}
		form.Add("url", "http://non-existent-domain.invalid")
		resp, err := http.PostForm(appServer.URL+"/analyze", form)
		if err != nil {
			t.Fatalf("failed to send request: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Could not reach the URL") {
			t.Error("response missing error message")
		}
	})
}
