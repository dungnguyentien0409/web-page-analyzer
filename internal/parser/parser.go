package parser

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type PageAnalysisRequest struct {
	HTML []byte
	URL  string
}

type PageAnalysisResult struct {
	HTMLVersion string
	Title       string

	Headings map[string]int

	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int

	ContainsLoginForm bool
}

type Analyzer interface {
	AnalyzePage(ctx context.Context, req PageAnalysisRequest) (*PageAnalysisResult, error)
}

type DefaultAnalyzer struct {
	documentProvider func(io.Reader) (*goquery.Document, error)
	logger           *slog.Logger
}

func NewDefaultAnalyzer(logger *slog.Logger) *DefaultAnalyzer {
	return &DefaultAnalyzer{
		logger:           logger,
		documentProvider: goquery.NewDocumentFromReader,
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
func (a *DefaultAnalyzer) AnalyzePage(ctx context.Context, req PageAnalysisRequest) (*PageAnalysisResult, error) {
	a.logger.Info("analyzing page", "url", req.URL)
	doc, err := a.ParseHTML(req.HTML)
	if err != nil {
		return nil, err
	}
	var (
		res = &PageAnalysisResult{}
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
		links, lErr := extractLinks(ctx, a.logger, doc, req.URL)
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
