package http

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics is the metrics for http probe
type Metrics struct {
	StatusCode *prometheus.CounterVec
	ContentLen *prometheus.GaugeVec
}

// NewMetrics create the HTTP metrics
func NewMetrics(subsystem, name string) *Metrics {
	namespace := global.GetEaseProbe().Name
	return &Metrics{
		StatusCode: metric.NewCounter(namespace, subsystem, name, "status_code",
			"HTTP Status Code", []string{"name", "status"}),
		ContentLen: metric.NewGauge(namespace, subsystem, name, "content_len",
			"HTTP Content Length", []string{"name", "status"}),
	}
}
