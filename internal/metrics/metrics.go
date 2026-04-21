package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Collector struct {
	httpRequests        *prometheus.CounterVec
	httpDuration        *prometheus.HistogramVec
	linksChecked        *prometheus.CounterVec
	outboundRequests    *prometheus.CounterVec
	outboundDuration    *prometheus.HistogramVec
	rateLimitRejections *prometheus.CounterVec
	inflightRequests    *prometheus.GaugeVec
}

var (
	instance *Collector
	once     sync.Once
)

var httpDurationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

var outboundDurationBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5}

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
				Buckets: httpDurationBuckets,
			}, []string{"status"}),
			linksChecked: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "web_analyzer_links_checked_total",
				Help: "Total number of links checked",
			}, []string{"accessible"}),
			outboundRequests: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "web_analyzer_outbound_requests_total",
				Help: "Total number of outbound requests to external domains",
			}, []string{"domain", "method", "status"}),
			outboundDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "web_analyzer_outbound_request_duration_seconds",
				Help:    "Duration of outbound HTTP requests in seconds",
				Buckets: outboundDurationBuckets,
			}, []string{"domain", "method"}),
			rateLimitRejections: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "web_analyzer_rate_limit_rejections_total",
				Help: "Total number of requests rejected by rate limiting",
			}, []string{"type"}),
			inflightRequests: promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "web_analyzer_inflight_requests",
				Help: "Number of requests currently being processed",
			}, []string{"endpoint"}),
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

func (c *Collector) ObserveOutboundDuration(domain, method string, duration float64) {
	c.outboundDuration.WithLabelValues(domain, method).Observe(duration)
}

func (c *Collector) IncRateLimitRejection(limitType string) {
	c.rateLimitRejections.WithLabelValues(limitType).Inc()
}

func (c *Collector) IncInflight(endpoint string) {
	c.inflightRequests.WithLabelValues(endpoint).Inc()
}

func (c *Collector) DecInflight(endpoint string) {
	c.inflightRequests.WithLabelValues(endpoint).Dec()
}
