package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesNewID(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check response header
	headerID := rr.Header().Get("X-Request-ID")
	if headerID == "" {
		t.Error("expected X-Request-ID header to be set")
	}

	// Check context value
	if capturedID == "" {
		t.Error("expected request ID in context")
	}

	// Both should match
	if headerID != capturedID {
		t.Errorf("header ID %q does not match context ID %q", headerID, capturedID)
	}
}

func TestRequestID_PreservesExistingID(t *testing.T) {
	existingID := "existing-request-id-123"
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", existingID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should preserve existing ID
	if capturedID != existingID {
		t.Errorf("expected ID %q, got %q", existingID, capturedID)
	}

	headerID := rr.Header().Get("X-Request-ID")
	if headerID != existingID {
		t.Errorf("expected header ID %q, got %q", existingID, headerID)
	}
}

func TestGetRequestID_EmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	id := GetRequestID(ctx)
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}