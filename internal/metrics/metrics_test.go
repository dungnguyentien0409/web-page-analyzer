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

	t.Run("IncAnalysisTotal", func(t *testing.T) {
		c.IncAnalysisTotal("success")
		c.IncAnalysisTotal("error")
	})
	t.Run("ObserveAnalysisDuration", func(t *testing.T) {
		c.ObserveAnalysisDuration("success", 0.123)
	})
	t.Run("IncLinksChecked", func(t *testing.T) {
		c.IncLinksChecked(true)
		c.IncLinksChecked(false)
	})
}
