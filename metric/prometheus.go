package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// // Metrics is the metrics type
// type Metrics int

// // ProbeMetrics metrics type
// const (
// 	Counter Metrics = iota
// 	Gauge
// 	Histogram
// 	Summary
// )

// MetricsType is the generic type of metrics
type MetricsType interface {
	*prometheus.CounterVec | *prometheus.GaugeVec | *prometheus.HistogramVec | *prometheus.SummaryVec
}

var counterMap = make(map[string]*prometheus.CounterVec)
var gaugeMap = make(map[string]*prometheus.GaugeVec)
var histogramMap = make(map[string]*prometheus.HistogramVec)
var summaryMap = make(map[string]*prometheus.SummaryVec)

// Counter get the counter metric by key
func Counter(key string) *prometheus.CounterVec {
	return counterMap[key]
}

// Gauge get the gauge metric by key
func Gauge(key string) *prometheus.GaugeVec {
	return gaugeMap[key]
}

// NewCounter create the counter metric
func NewCounter(namespace, subsystem, name, metric string,
	help string, labels []string) *prometheus.CounterVec {

	metricName := GetName(namespace, subsystem, name, metric)
	if m, find := counterMap[metricName]; find {
		return m
	}

	counterMap[metricName] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: metricName,
			Help: help,
		},
		labels,
	)
	prometheus.MustRegister()
	return counterMap[metricName]
}

// NewGauge create the gauge metric
func NewGauge(namespace, subsystem, name, metric string,
	help string, labels []string) *prometheus.GaugeVec {

	metricName := GetName(namespace, subsystem, name, metric)
	if m, find := gaugeMap[metricName]; find {
		return m
	}

	gaugeMap[metricName] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricName,
			Help: help,
		},
		labels,
	)
	prometheus.MustRegister(gaugeMap[metricName])
	return gaugeMap[metricName]
}

// GetName generate the metric key by a number of strings
func GetName(fields ...string) string {
	key := ""
	for _, v := range fields {
		if len(v) > 0 {
			key += RemoveInvalidChars(v) + "_"
		}
	}

	if len(key) > 0 && key[len(key)-1] == '_' {
		key = key[:len(key)-1]
	}

	log.Debugf("metric key: %s", key)
	return key
}

// ValidName check if the char is valid
func ValidName(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
}

// RemoveInvalidChars remove invalid chars
func RemoveInvalidChars(name string) string {
	var result []byte
	for i := 0; i < len(name); i++ {
		if ValidName(name[i]) {
			result = append(result, name[i])
		}

	}
	return string(result)
}
