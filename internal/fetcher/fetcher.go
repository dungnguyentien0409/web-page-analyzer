package fetcher

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/metrics"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

type FetcherConfig struct {
	TimeoutSec          int
	DialTimeoutSec      int
	DialKeepAliveSec    int
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeoutSec  int
	TLSHandshakeSec     int
}

type DefaultFetcher struct {
	client  *http.Client
	logger  *slog.Logger
	metrics *metrics.Collector
}

func NewDefaultFetcher(cfg FetcherConfig, logger *slog.Logger, metrics *metrics.Collector) *DefaultFetcher {
	// Apply defaults
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 10
	}
	if cfg.DialTimeoutSec <= 0 {
		cfg.DialTimeoutSec = 5
	}
	if cfg.DialKeepAliveSec <= 0 {
		cfg.DialKeepAliveSec = 30
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 100
	}
	if cfg.MaxIdleConnsPerHost <= 0 {
		cfg.MaxIdleConnsPerHost = 20
	}
	if cfg.IdleConnTimeoutSec <= 0 {
		cfg.IdleConnTimeoutSec = 90
	}
	if cfg.TLSHandshakeSec <= 0 {
		cfg.TLSHandshakeSec = 5
	}

	return &DefaultFetcher{
		logger:  logger,
		metrics: metrics,
		client: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSec) * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   time.Duration(cfg.DialTimeoutSec) * time.Second,
					KeepAlive: time.Duration(cfg.DialKeepAliveSec) * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          cfg.MaxIdleConns,
				MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
				IdleConnTimeout:       time.Duration(cfg.IdleConnTimeoutSec) * time.Second,
				TLSHandshakeTimeout:   time.Duration(cfg.TLSHandshakeSec) * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}
func (f *DefaultFetcher) Fetch(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "InsightBot/1.0 (Web Page Analyzer)")
	f.logger.Info("fetching url", "url", urlStr)

	u, _ := url.Parse(urlStr)
	domain := "unknown"
	if u != nil {
		domain = u.Host
	}

	start := time.Now()
	resp, err := f.client.Do(req)
	duration := time.Since(start).Seconds()

	status := "error"
	if resp != nil {
		status = fmt.Sprintf("%d", resp.StatusCode)
	}
	f.metrics.IncOutboundRequest(domain, "GET", status)
	f.metrics.ObserveOutboundDuration(domain, "GET", duration)

	if err != nil {
		f.logger.Error("fetch failed", "url", urlStr, "error", err)
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		f.logger.Warn("bad status code", "url", urlStr, "status", resp.StatusCode)
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
