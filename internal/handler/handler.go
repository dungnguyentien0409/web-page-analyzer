package handler

import (
	"bytes"
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/analyzer"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/middleware"
)

type HandlerConfig struct {
	Template       *template.Template
	Fetcher        fetcher.Fetcher
	Analyzer       analyzer.Analyzer
	RequestTimeout time.Duration
	Logger         *slog.Logger
	Metrics        *metrics.Collector
}

type Handler struct {
	tmpl           *template.Template
	fetcher        fetcher.Fetcher
	analyzePage    func(context.Context, analyzer.AnalysisRequest) (*analyzer.AnalysisResult, error)
	requestTimeout time.Duration
	logger         *slog.Logger
	metrics        *metrics.Collector
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		tmpl:           cfg.Template,
		fetcher:        cfg.Fetcher,
		analyzePage:    cfg.Analyzer.AnalyzePage,
		requestTimeout: cfg.RequestTimeout,
		logger:         cfg.Logger,
		metrics:        cfg.Metrics,
	}
}
func (h *Handler) render(w http.ResponseWriter, data any) {
	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		h.logger.Error("failed to write response", "error", err)
	}
}
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	h.render(w, nil)
}
func (h *Handler) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With("request_id", requestID)

	urlInput := r.FormValue("url")
	if urlInput == "" {
		logger.Warn("missing url parameter")
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	logger.Info("analysis request started", "url", urlInput)
	start := time.Now()
	status := "success"
	defer func() {
		h.metrics.IncHTTPRequests(status)
		h.metrics.ObserveHTTPDuration(status, time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
	defer cancel()

	htmlContent, err := h.fetcher.Fetch(ctx, urlInput)
	if err != nil {
		status = "fetch_error"
		logger.Error("fetch error", "url", urlInput, "error", err)
		h.render(w, map[string]any{"Error": "Could not reach the URL. " + err.Error()})
		return
	}
	analysis, err := h.analyzePage(ctx, analyzer.AnalysisRequest{HTML: htmlContent, URL: urlInput})
	if err != nil {
		status = "analysis_error"
		logger.Error("analysis error", "url", urlInput, "error", err)
		h.render(w, map[string]any{"Error": "Failed to parse HTML"})
		return
	}

	logger.Info("analysis completed",
		"url", urlInput,
		"title", analysis.Title,
		"internal_links", analysis.InternalLinks,
		"external_links", analysis.ExternalLinks,
		"inaccessible_links", analysis.InaccessibleLinks,
		"duration", time.Since(start).String(),
	)

	h.render(w, map[string]any{"Result": analysis})
}
