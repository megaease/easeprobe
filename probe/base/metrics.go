package base

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics is the probe metrics
type Metrics struct {
	Total    *prometheus.CounterVec
	Duration *prometheus.GaugeVec
	Status   *prometheus.GaugeVec
}

// NewMetrics create the metrics
func NewMetrics(subsystem, name string) *Metrics {
	namespace := global.GetEaseProbe().Name
	return &Metrics{
		Total: metric.NewCounter(namespace, subsystem, name, "total",
			"Total Probe Number", []string{"name", "status"}),
		Duration: metric.NewGauge(namespace, subsystem, name, "duration",
			"Probe Duration", []string{"name", "status"}),
		Status: metric.NewGauge(namespace, subsystem, name, "status",
			"Probe Status", []string{"name"}),
	}
}
