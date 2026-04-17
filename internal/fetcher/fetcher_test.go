package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchURL_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		}),
	)
	defer server.Close()

	body, err := FetchURL(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "test response"

	if string(body) != expected {
		t.Errorf(
			"expected body %q, got %q",
			expected,
			string(body),
		)
	}
}

func TestFetchURL_HTTPError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	defer server.Close()

	_, err := FetchURL(server.URL)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchURL_InvalidURL(t *testing.T) {
	_, err := FetchURL("http://invalid-url")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}