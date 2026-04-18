package handler

import (
	"html/template"
	"net/http"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

type Handler struct {
	tmpl        *template.Template
	fetchURL    func(string) ([]byte, error)
	analyzePage func(parser.PageSource) (*parser.PageAnalysis, error)
}

func NewHandler(t *template.Template, a *parser.Analyzer) *Handler {
	return &Handler{
		tmpl:        t,
		fetchURL:    fetcher.FetchURL,
		analyzePage: a.AnalyzePage,
	}
}
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	htmlContent, err := h.fetchURL(urlInput)
	if err != nil {
		h.tmpl.Execute(w, map[string]any{"Error": "Could not reach the URL. " + err.Error()})
		return
	}
	analysis, err := h.analyzePage(parser.PageSource{HTML: htmlContent, URL: urlInput})
	if err != nil {
		h.tmpl.Execute(w, map[string]any{"Error": "Failed to parse HTML"})
		return
	}
	err = h.tmpl.Execute(w, map[string]any{
		"Result": analysis,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
