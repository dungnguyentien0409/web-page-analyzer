package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Fetcher interface {
	Fetch(url string) ([]byte, error)
}
type DefaultFetcher struct {
	Timeout time.Duration
}

func NewDefaultFetcher() *DefaultFetcher {
	return &DefaultFetcher{
		Timeout: 5 * time.Second,
	}
}
func (f *DefaultFetcher) Fetch(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: f.Timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
