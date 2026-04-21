package analyzer

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/ratelimit"

	"github.com/PuerkitoBio/goquery"
)

type AnalysisRequest struct {
	HTML []byte
	URL  string
}

type AnalysisResult struct {
	HTMLVersion string
	Title       string

	Headings map[string]int

	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int

	ContainsLoginForm bool
}

type Analyzer interface {
	AnalyzePage(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error)
}

type AnalyzerConfig struct {
	Logger           *slog.Logger
	RetryCount       int
	WorkerCount      int
	Metrics          *metrics.Collector
	OutboundLimiter  *ratelimit.OutboundLimiter
}

type DefaultAnalyzer struct {
	documentProvider func(io.Reader) (*goquery.Document, error)
	logger           *slog.Logger
	retryCount       int
	workerCount      int
	metrics          *metrics.Collector
	outboundLimiter  *ratelimit.OutboundLimiter
}

func NewDefaultAnalyzer(cfg AnalyzerConfig) *DefaultAnalyzer {
	return &DefaultAnalyzer{
		logger:           cfg.Logger,
		documentProvider: goquery.NewDocumentFromReader,
		retryCount:       cfg.RetryCount,
		workerCount:      cfg.WorkerCount,
		metrics:          cfg.Metrics,
		outboundLimiter:  cfg.OutboundLimiter,
	}
}
func (a *DefaultAnalyzer) ParseHTML(html []byte) (*goquery.Document, error) {
	reader := bytes.NewReader(html)
	doc, err := a.documentProvider(reader)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
func (a *DefaultAnalyzer) AnalyzePage(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	a.logger.Info("analyzing page", "url", req.URL)
	doc, err := a.ParseHTML(req.HTML)
	if err != nil {
		return nil, err
	}
	var (
		res = &AnalysisResult{}
		wg  sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		res.HTMLVersion = detectHTMLVersion(req.HTML)
		res.Title = ExtractTitle(doc)
		res.Headings = countHeadings(doc)
		res.ContainsLoginForm = checkLoginForm(doc)
	}()
	go func() {
		defer wg.Done()
		links, lErr := a.extractLinks(ctx, doc, req.URL)
		if lErr != nil {
			err = lErr
			return
		}
		res.InternalLinks = links.Internal
		res.ExternalLinks = links.External
		res.InaccessibleLinks = links.Inaccessible
	}()
	wg.Wait()
	return res, err
}
