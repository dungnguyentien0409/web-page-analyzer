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
type DefaultFetcher struct {
	client  *http.Client
	logger  *slog.Logger
	metrics *metrics.Collector
}

func NewDefaultFetcher(logger *slog.Logger, metrics *metrics.Collector) *DefaultFetcher {
	return &DefaultFetcher{
		logger:  logger,
		metrics: metrics,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
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

	resp, err := f.client.Do(req)

	status := "error"
	if resp != nil {
		status = fmt.Sprintf("%d", resp.StatusCode)
	}
	f.metrics.IncOutboundRequest(domain, "GET", status)

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
