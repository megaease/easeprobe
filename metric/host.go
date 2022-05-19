package metric

import "github.com/prometheus/client_golang/prometheus"

// HostMetrics is the metrics for host probe
type HostMetrics struct {
	CPU    *prometheus.GaugeVec
	Memory *prometheus.GaugeVec
	Disk   *prometheus.GaugeVec
}

// NewHostMetrics create the host metrics
func NewHostMetrics(namespace, subsystem, name string) *HostMetrics {

	return &HostMetrics{
		CPU: NewGauge(namespace, subsystem, name, "cpu",
			"CPU Usage", []string{"host", "state"}),
		Memory: NewGauge(namespace, subsystem, name, "memory",
			"Memory Usage", []string{"host", "state"}),
		Disk: NewGauge(namespace, subsystem, name, "disk",
			"Disk Usage", []string{"host", "state"}),
	}
}
