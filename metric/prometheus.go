/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package metric is the package to report the metrics to Prometheus
package metric

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const module = "Metric"

// MetricsType is the generic type of metrics
type MetricsType interface {
	*prometheus.CounterVec | *prometheus.GaugeVec | *prometheus.HistogramVec | *prometheus.SummaryVec
}

var (
	counterMap   = make(map[string]*prometheus.CounterVec)
	gaugeMap     = make(map[string]*prometheus.GaugeVec)
	histogramMap = make(map[string]*prometheus.HistogramVec)
	summaryMap   = make(map[string]*prometheus.SummaryVec)
)

var (
	validMetric = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
	validLabel  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

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

	metricName, err := getAndValid(namespace, subsystem, name, metric, labels)
	if err != nil {
		log.Errorf("[%s] %v", module, err)
		return nil
	}

	if m, find := counterMap[metricName]; find {
		log.Debugf("[%s] Counter <%s> already created!", module, metricName)
		return m
	}

	counterMap[metricName] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: metricName,
			Help: help,
		},
		labels,
	)
	prometheus.MustRegister(counterMap[metricName])

	log.Infof("[%s] Counter <%s> is created!", module, metricName)
	return counterMap[metricName]
}

// NewGauge create the gauge metric
func NewGauge(namespace, subsystem, name, metric string,
	help string, labels []string) *prometheus.GaugeVec {

	metricName, err := getAndValid(namespace, subsystem, name, metric, labels)
	if err != nil {
		log.Errorf("[%s] %v", module, err)
		return nil
	}
	if m, find := gaugeMap[metricName]; find {
		log.Debugf("[%s] Gauge <%s> already created!", module, metricName)
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

	log.Infof("[%s] Gauge <%s> is created!", module, metricName)
	return gaugeMap[metricName]
}

func getAndValid(namespace, subsystem, name, metric string, labels []string) (string, error) {
	metricName := GetName(namespace, subsystem, name, metric)
	if ValidMetricName(metricName) == false {
		return "", fmt.Errorf("Invalid metric name: %s", metricName)
	}

	for _, l := range labels {
		if ValidLabelName(l) == false {
			return "", fmt.Errorf("Invalid label name: %s", l)
		}
	}
	return metricName, nil
}

// GetName generate the metric key by a number of strings
func GetName(fields ...string) string {
	name := ""
	for _, v := range fields {
		v = RemoveInvalidChars(v)
		if len(v) > 0 {
			name += v + "_"
		}
	}

	if len(name) > 0 && name[len(name)-1] == '_' {
		name = name[:len(name)-1]
	}

	log.Debugf("[%s] get the name: %s", module, name)
	return name
}

// ValidMetricName check if the metric name is valid
func ValidMetricName(name string) bool {
	return validMetric.MatchString(name)
}

// ValidLabelName check if the label name is valid
func ValidLabelName(label string) bool {
	return validLabel.MatchString(label)
}

// ValidMetricChar check if the char is valid for metric name
func ValidMetricChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_' || ch == ':'
}

// RemoveInvalidChars remove invalid chars
func RemoveInvalidChars(name string) string {
	var result []byte
	i := 0

	// skip all of the non-alphabetic chars
	for ; i < len(name); i++ {
		ch := name[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			break
		}
	}

	// remove the invalid chars
	for ; i < len(name); i++ {
		if ValidMetricChar(name[i]) {
			result = append(result, name[i])
		}
	}
	return string(result)
}
