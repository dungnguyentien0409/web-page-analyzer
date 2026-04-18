package handler

import (
	"bytes"
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

type HandlerConfig struct {
	Template       *template.Template
	Fetcher        fetcher.Fetcher
	Analyzer       parser.Analyzer
	RequestTimeout time.Duration
	Logger         *slog.Logger
}

type Handler struct {
	tmpl           *template.Template
	fetcher        fetcher.Fetcher
	analyzePage    func(context.Context, parser.PageAnalysisRequest) (*parser.PageAnalysisResult, error)
	requestTimeout time.Duration
	logger         *slog.Logger
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		tmpl:           cfg.Template,
		fetcher:        cfg.Fetcher,
		analyzePage:    cfg.Analyzer.AnalyzePage,
		requestTimeout: cfg.RequestTimeout,
		logger:         cfg.Logger,
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
	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	htmlContent, err := h.fetcher.Fetch(ctx, urlInput)
	if err != nil {
		h.logger.Error("fetch error in handler", "url", urlInput, "error", err)
		h.render(w, map[string]any{"Error": "Could not reach the URL. " + err.Error()})
		return
	}
	analysis, err := h.analyzePage(ctx, parser.PageAnalysisRequest{HTML: htmlContent, URL: urlInput})
	if err != nil {
		h.logger.Error("analysis error in handler", "url", urlInput, "error", err)
		h.render(w, map[string]any{"Error": "Failed to parse HTML"})
		return
	}
	h.render(w, map[string]any{"Result": analysis})
}
