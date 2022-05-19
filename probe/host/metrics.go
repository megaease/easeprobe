package host

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics is the metrics for host probe
type Metrics struct {
	CPU    *prometheus.GaugeVec
	Memory *prometheus.GaugeVec
	Disk   *prometheus.GaugeVec
}

// NewMetrics create the host metrics
func NewMetrics(subsystem, name string) *Metrics {
	namespace := global.GetEaseProbe().Name
	return &Metrics{
		CPU: metric.NewGauge(namespace, subsystem, name, "cpu",
			"CPU Usage", []string{"host", "state"}),
		Memory: metric.NewGauge(namespace, subsystem, name, "memory",
			"Memory Usage", []string{"host", "state"}),
		Disk: metric.NewGauge(namespace, subsystem, name, "disk",
			"Disk Usage", []string{"host", "state"}),
	}
}
