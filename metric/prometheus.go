package metric

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Metrics is the probe metrics
type Metrics struct {
	Total    *prometheus.CounterVec
	Duration *prometheus.GaugeVec
	Status   *prometheus.GaugeVec
}

var metrics = make(map[string]*Metrics)

// NewMetrics create the metrics
func NewMetrics(namespace, subsystem, name string) *Metrics {

	key := metricName(namespace) + "_" + metricName(subsystem)
	if name != "" {
		key = key + "_" + metricName(name)
	}

	log.Debugf("metric key: %s", key)

	if m, find := metrics[key]; find {
		return m
	}

	m := &Metrics{
		Total: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        key + "_total",
				Help:        "Total Probe Number",
				ConstLabels: map[string]string{},
			},
			[]string{"probe", "status"},
		),
		Duration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      key + "_duration",
				Help:      "Duration of Probe",
			},
			[]string{"probe"},
		),
		Status: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      key + "_status",
				Help:      "Status of Probe",
			},
			[]string{"probe"},
		),
	}

	prometheus.MustRegister(m.Total)
	prometheus.MustRegister(m.Duration)
	prometheus.MustRegister(m.Status)

	metrics[key] = m
	return m
}

func valid(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
}

func metricName(name string) string {
	metricName := []byte(strings.TrimSpace(name))
	for i := 0; i < len(name); i++ {
		if valid(name[i]) {
			continue
		}
		metricName[i] = ' '
	}
	words := strings.Fields(string(metricName))
	return strings.Join(words, "_")
}
