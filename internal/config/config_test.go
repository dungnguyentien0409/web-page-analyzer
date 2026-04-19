package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		content := `{"request_timeout_seconds": 10, "link_check_workers": 5, "link_check_retries": 2}`
		tmpFile, err := os.CreateTemp("", "config_*.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		_ = tmpFile.Close() // Close the file after writing
		cfg, err := Load(tmpFile.Name())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg.RequestTimeoutSeconds != 10 || cfg.LinkCheckWorkers != 5 || cfg.LinkCheckRetries != 2 {
			t.Errorf("unexpected config values: %+v", cfg)
		}
	})
	t.Run("FileNotFound", func(t *testing.T) {
		_, err := Load("non_existent.json")
		if err == nil {
			t.Fatal("expected error for non-existent file")
		}
	})
	t.Run("InvalidJSON", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "invalid_*.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		if _, err := tmpFile.Write([]byte(`{invalid}`)); err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		_ = tmpFile.Close()
		_, err = Load(tmpFile.Name())
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}
