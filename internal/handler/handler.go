package handler

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

type Handler struct {
	tmpl        *template.Template
	fetcher     fetcher.Fetcher
	analyzePage func(parser.PageAnalysisRequest) (*parser.PageAnalysisResult, error)
}

func NewHandler(t *template.Template, f fetcher.Fetcher, a parser.Analyzer) *Handler {
	return &Handler{
		tmpl:        t,
		fetcher:     f,
		analyzePage: a.AnalyzePage,
	}
}
func (h *Handler) render(w http.ResponseWriter, data any) {
	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	h.render(w, nil)
}
func (h *Handler) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	urlInput := r.FormValue("url")
	if urlInput == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}
	htmlContent, err := h.fetcher.Fetch(urlInput)
	if err != nil {
		h.render(w, map[string]any{"Error": "Could not reach the URL. " + err.Error()})
		return
	}
	analysis, err := h.analyzePage(parser.PageAnalysisRequest{HTML: htmlContent, URL: urlInput})
	if err != nil {
		h.render(w, map[string]any{"Error": "Failed to parse HTML"})
		return
	}
	h.render(w, map[string]any{"Result": analysis})
}
