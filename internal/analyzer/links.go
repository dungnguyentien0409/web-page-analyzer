package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type LinkAnalysisResult struct {
	Internal     int
	External     int
	Inaccessible int
}

var (
	linkCheckClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
		},
	}
)

func (a *DefaultAnalyzer) isLinkAccessible(ctx context.Context, link string) bool {
	u, _ := url.Parse(link)
	domain := "unknown"
	if u != nil {
		domain = u.Host
	}

	// Rate limit trước khi request
	if a.outboundLimiter != nil {
		if err := a.outboundLimiter.Wait(ctx, domain); err != nil {
			a.metrics.IncRateLimitRejection("outbound")
			return false
		}
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", link, nil)
	if err != nil {
		return false
	}

	a.logger.Debug("checking link", "url", link)
	var resp *http.Response
	for i := 0; i < a.retryCount; i++ {
		reqStart := time.Now()
		resp, err = linkCheckClient.Do(req)
		reqDuration := time.Since(reqStart).Seconds()

		status := "error"
		if resp != nil {
			status = fmt.Sprintf("%d", resp.StatusCode)
		}
		a.metrics.IncOutboundRequest(domain, "HEAD", status)
		a.metrics.ObserveOutboundDuration(domain, "HEAD", reqDuration)

		if err == nil && resp.StatusCode < 500 {
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		if ctx.Err() != nil {
			return false
		}
		a.logger.Warn("retrying link check", "url", link, "attempt", i+1)
	}

	accessible := false
	if err == nil && resp != nil {
		defer func() { _ = resp.Body.Close() }()
		accessible = resp.StatusCode >= 200 && resp.StatusCode < 400
	}

	a.metrics.IncLinksChecked(accessible)
	return accessible
}

func (a *DefaultAnalyzer) extractLinks(
	ctx context.Context,
	doc *goquery.Document,
	baseURLStr string,
) (*LinkAnalysisResult, error) {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, err
	}
	res := &LinkAnalysisResult{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	linkSet := make(map[string]bool)
	jobs := make(chan string)
	for w := 0; w < a.workerCount; w++ {
		go func() {
			for l := range jobs {
				if !a.isLinkAccessible(ctx, l) {
					mu.Lock()
					res.Inaccessible++
					mu.Unlock()
				}
				wg.Done()
			}
		}()
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		if href == "" ||
			strings.HasPrefix(href, "#") ||
			strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "javascript:") {
			return
		}

		linkURL, err := url.Parse(href)
		if err != nil {
			return
		}

		resolved := baseURL.ResolveReference(linkURL)
		if resolved.Host == "" {
			return
		}
		fullURL := resolved.String()
		if !linkSet[fullURL] {
			linkSet[fullURL] = true
			if resolved.Host == baseURL.Host {
				res.Internal++
			} else {
				res.External++
			}
			wg.Add(1)
			jobs <- fullURL
		}
	})
	close(jobs)
	wg.Wait()
	return res, nil
}
