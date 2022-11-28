package host

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// IMetrics is the interface of metrics
type IMetrics interface {
	Command() string                                   // Command returns the command to get the metrics
	OutputLines() int                                  // OutputLines returns the lines of command output
	Config(s *Server) error                            // Config returns the config of the metrics
	SetThreshold(t *Threshold)                         // SetThreshold sets the threshold of the metrics
	Parse(s []string) error                            // Parse a string to a metrics struct
	UsageInfo() string                                 // UsageInfo returns the usage info of the metrics
	CheckThreshold() (bool, string)                    // CheckThreshold check the metrics usage
	ExportMetrics(name string, g *prometheus.GaugeVec) // ExportMetrics export the metrics
}

// ResourceUsage is the resource usage for cpu and memory
type ResourceUsage struct {
	Used  int     `yaml:"used"`
	Total int     `yaml:"total"`
	Usage float64 `yaml:"usage"`
	Tag   string  `yaml:"tag"`
}

func first(str string) string {
	return strings.Split(strings.TrimSpace(str), " ")[0]
}

func strFloat(str string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(str), 32)
	return n
}

func strInt(str string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(str), 10, 32)
	return n
}

func addMessage(msg string, message string) string {
	if msg == "" || message == "" {
		return message
	}
	return msg + " | " + message
}
