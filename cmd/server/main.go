package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/config"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/handler"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
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
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	fetcherSvc := fetcher.NewDefaultFetcher(logger)
	analyzerSvc := parser.NewDefaultAnalyzer(parser.AnalyzerConfig{
		Logger:      logger,
		RetryCount:  cfg.LinkCheckRetries,
		WorkerCount: cfg.LinkCheckWorkers,
	})
	h := handler.NewHandler(handler.HandlerConfig{
		Template:       tmpl,
		Fetcher:        fetcherSvc,
		Analyzer:       analyzerSvc,
		RequestTimeout: time.Duration(cfg.RequestTimeoutSeconds) * time.Second,
		Logger:         logger,
	})
	http.HandleFunc("/", h.IndexHandler)
	http.HandleFunc("/analyze", h.AnalyzeHandler)
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("web/static")),
		),
	)
	srv := &http.Server{
		Addr: ":8080",
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
	logger.Info("Server exiting")
}
