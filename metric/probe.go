package metric

import "github.com/prometheus/client_golang/prometheus"

// ProbeMetrics is the probe metrics
type ProbeMetrics struct {
	Total    *prometheus.CounterVec
	Duration *prometheus.GaugeVec
	Status   *prometheus.GaugeVec
}

// NewProbeMetrics create the metrics
func NewProbeMetrics(namespace, subsystem, name string) *ProbeMetrics {
	return &ProbeMetrics{
		Total: NewCounter(namespace, subsystem, name, "total",
			"Total Probe Number", []string{"name", "status"}),
		Duration: NewGauge(namespace, subsystem, name, "duration",
			"Probe Duration", []string{"name", "status"}),
		Status: NewGauge(namespace, subsystem, name, "status",
			"Probe Status", []string{"name"}),
	}
}
