package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	RequestTimeoutSeconds int `json:"request_timeout_seconds"`
	LinkCheckWorkers      int `json:"link_check_workers"`
	LinkCheckRetries      int `json:"link_check_retries"`
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
