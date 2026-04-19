package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Collector struct {
	analysisTotal    *prometheus.CounterVec
	analysisDuration *prometheus.HistogramVec
	linksChecked     *prometheus.CounterVec
}

func NewCollector() *Collector {
	return &Collector{
		analysisTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "web_analyzer_analysis_requests_total",
			Help: "Total number of page analysis requests",
		}, []string{"status"}),
		analysisDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "web_analyzer_analysis_duration_seconds",
			Help:    "Duration of page analysis in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"status"}),
		linksChecked: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "web_analyzer_links_checked_total",
			Help: "Total number of links checked",
		}, []string{"accessible"}),
	}
}

func (c *Collector) IncAnalysisTotal(status string) {
	c.analysisTotal.WithLabelValues(status).Inc()
}

func (c *Collector) ObserveAnalysisDuration(status string, duration float64) {
	c.analysisDuration.WithLabelValues(status).Observe(duration)
}

func (c *Collector) IncLinksChecked(accessible bool) {
	val := "true"
	if !accessible {
		val = "false"
	}
	c.linksChecked.WithLabelValues(val).Inc()
}
