package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	RequestTimeoutSeconds int     `json:"request_timeout_seconds"`
	LinkCheckWorkers      int     `json:"link_check_workers"`
	LinkCheckRetries      int     `json:"link_check_retries"`
	RateLimitRPS          float64 `json:"rate_limit_rps"`
	RateLimitBurst        int     `json:"rate_limit_burst"`

	// Outbound rate limiting
	OutboundGlobalRPS   int `json:"outbound_global_rps"`
	OutboundGlobalBurst int `json:"outbound_global_burst"`
	OutboundHostRPS     int `json:"outbound_host_rps"`
	OutboundHostBurst   int `json:"outbound_host_burst"`

	// Fetcher HTTP client settings
	FetcherTimeoutSec         int `json:"fetcher_timeout_sec"`
	FetcherDialTimeoutSec     int `json:"fetcher_dial_timeout_sec"`
	FetcherDialKeepAliveSec   int `json:"fetcher_dial_keep_alive_sec"`
	FetcherMaxIdleConns       int `json:"fetcher_max_idle_conns"`
	FetcherMaxIdleConnsPerHost int `json:"fetcher_max_idle_conns_per_host"`
	FetcherIdleConnTimeoutSec int `json:"fetcher_idle_conn_timeout_sec"`
	FetcherTLSHandshakeSec    int `json:"fetcher_tls_handshake_sec"`

	// Link checker HTTP client settings
	LinkCheckTimeoutSec       int `json:"link_check_timeout_sec"`
	LinkCheckMaxIdleConns     int `json:"link_check_max_idle_conns"`
	LinkCheckMaxIdlePerHost   int `json:"link_check_max_idle_per_host"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
