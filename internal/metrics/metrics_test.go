package metrics

import (
	"testing"
)

func TestCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("expected collector instance, got nil")
	}

	c2 := NewCollector()
	if c != c2 {
		t.Error("NewCollector should return the same instance (singleton)")
	}
	t.Run("IncHTTPRequests", func(t *testing.T) {
		c.IncHTTPRequests("success")
		c.IncHTTPRequests("error")
	})
	t.Run("ObserveHTTPDuration", func(t *testing.T) {
		c.ObserveHTTPDuration("success", 0.123)
	})
	t.Run("IncLinksChecked", func(t *testing.T) {
		c.IncLinksChecked(true)
		c.IncLinksChecked(false)
	})
	t.Run("IncOutboundRequest", func(t *testing.T) {
		c.IncOutboundRequest("example.com", "GET", "200")
		c.IncOutboundRequest("broken.com", "HEAD", "error")
	})
}
