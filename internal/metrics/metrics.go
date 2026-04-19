package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Collector struct {
	httpRequests     *prometheus.CounterVec
	httpDuration     *prometheus.HistogramVec
	linksChecked     *prometheus.CounterVec
	outboundRequests *prometheus.CounterVec
}

var (
	instance *Collector
	once     sync.Once
)

func NewCollector() *Collector {
	once.Do(func() {
		instance = &Collector{
			httpRequests: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "web_analyzer_http_requests_total",
				Help: "Total number of HTTP requests processed by the server",
			}, []string{"status"}),
			httpDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "web_analyzer_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			}, []string{"status"}),
			linksChecked: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "web_analyzer_links_checked_total",
				Help: "Total number of links checked",
			}, []string{"accessible"}),
			outboundRequests: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "web_analyzer_outbound_requests_total",
				Help: "Total number of outbound requests to external domains",
			}, []string{"domain", "method", "status"}),
		}
	})
	return instance
}

func (c *Collector) IncHTTPRequests(status string) {
	c.httpRequests.WithLabelValues(status).Inc()
}

func (c *Collector) ObserveHTTPDuration(status string, duration float64) {
	c.httpDuration.WithLabelValues(status).Observe(duration)
}

func (c *Collector) IncLinksChecked(accessible bool) {
	val := "false"
	if accessible {
		val = "true"
	}
	c.linksChecked.WithLabelValues(val).Inc()
}

func (c *Collector) IncOutboundRequest(domain, method, status string) {
	c.outboundRequests.WithLabelValues(domain, method, status).Inc()
}
