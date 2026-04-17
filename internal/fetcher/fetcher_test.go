package fetcher

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func (e errorReader) Close() error {
	return nil
}

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
		t.Errorf("expected body %q, got %q", expected, string(body))
	}
}

func TestFetchURL_HTTPErrorStatus(t *testing.T) {
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
	_, err := FetchURL("://bad-url")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchURL_ReadError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)

			hijacker, ok := w.(http.Hijacker)

			if !ok {
				t.Fatal("hijacking not supported")
			}

			conn, _, err := hijacker.Hijack()

			if err != nil {
				t.Fatal(err)
			}

			conn.Close()
		}),
	)

	defer server.Close()

	_, err := FetchURL(server.URL)

	if err == nil {
		t.Fatal("expected read error")
	}
}
