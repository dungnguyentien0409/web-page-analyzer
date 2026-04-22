package main

import (
	"context"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/analyzer"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/config"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/handler"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/middleware"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/ratelimit"
	"github.com/dungnguyentien0409/web-page-analyzer/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/development.json"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	mc := metrics.NewCollector()

	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		logger.Error("failed to sub templates", "error", err)
		os.Exit(1)
	}
	tmpl := template.Must(template.ParseFS(tmplFS, "index.html"))

	fetcherSvc := fetcher.NewDefaultFetcher(fetcher.FetcherConfig{
		TimeoutSec:          cfg.FetcherTimeoutSec,
		DialTimeoutSec:      cfg.FetcherDialTimeoutSec,
		DialKeepAliveSec:    cfg.FetcherDialKeepAliveSec,
		MaxIdleConns:        cfg.FetcherMaxIdleConns,
		MaxIdleConnsPerHost: cfg.FetcherMaxIdleConnsPerHost,
		IdleConnTimeoutSec:  cfg.FetcherIdleConnTimeoutSec,
		TLSHandshakeSec:     cfg.FetcherTLSHandshakeSec,
	}, logger, mc)

	outboundLimiter := ratelimit.NewOutboundLimiter(ratelimit.OutboundConfig{
		GlobalRPS:   cfg.OutboundGlobalRPS,
		GlobalBurst: cfg.OutboundGlobalBurst,
		HostRPS:     cfg.OutboundHostRPS,
		HostBurst:   cfg.OutboundHostBurst,
		Logger:      logger,
	})

	analyzerSvc := analyzer.NewDefaultAnalyzer(analyzer.AnalyzerConfig{
		Logger:          logger,
		RetryCount:      cfg.LinkCheckRetries,
		WorkerCount:     cfg.LinkCheckWorkers,
		Metrics:         mc,
		OutboundLimiter: outboundLimiter,
		LinkCheckClient: &analyzer.LinkCheckConfig{
			TimeoutSec:          cfg.LinkCheckTimeoutSec,
			MaxIdleConns:        cfg.LinkCheckMaxIdleConns,
			MaxIdleConnsPerHost: cfg.LinkCheckMaxIdlePerHost,
		},
	})
	h := handler.NewHandler(handler.HandlerConfig{
		Template:       tmpl,
		Fetcher:        fetcherSvc,
		Analyzer:       analyzerSvc,
		RequestTimeout: time.Duration(cfg.RequestTimeoutSeconds) * time.Second,
		Logger:         logger,
		Metrics:        mc,
	})

	limiter := ratelimit.NewInboundLimiter(ratelimit.InboundConfig{
		RPS:     cfg.RateLimitRPS,
		Burst:   cfg.RateLimitBurst,
		Metrics: mc,
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.IndexHandler)
	mux.HandleFunc("/analyze", h.AnalyzeHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("web/static")),
		),
	)

	// Middleware chain: RequestID -> RateLimit -> Handlers
	handler := middleware.RequestID(limiter.Middleware(mux))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	go func() {
		logger.Info("Server running", "addr", ":8080", "url", "http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}
	outboundLimiter.Stop()
	logger.Info("Server exiting")
}
