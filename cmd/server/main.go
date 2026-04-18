package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/handler"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

func main() {
	const requestTimeout = 15 * time.Second
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	fetcherSvc := fetcher.NewDefaultFetcher()
	analyzer := parser.NewDefaultAnalyzer()
	h := handler.NewHandler(handler.HandlerConfig{
		Template:       tmpl,
		Fetcher:        fetcherSvc,
		Analyzer:       analyzer,
		RequestTimeout: requestTimeout,
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
		log.Println("Server running at http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
