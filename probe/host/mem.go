package host

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Mem is the resource usage for memory and disk
type Mem struct {
	ResourceUsage `yaml:",inline"`
	Threshold     float64 `yaml:"threshold"`
}

// Command returns the command to get the cpu usage
func (m *Mem) Command() string {
	return `free -m | awk 'NR==2{printf "%s %s %.2f\n", $3,$2,$3*100/$2 }'`
}

// OutputLines returns the lines of command output
func (m *Mem) OutputLines() int {
	return 1
}

// Config returns the config of the cpu
func (m *Mem) Config(s *Server) error {
	if s.Threshold.Mem == 0 {
		s.Threshold.Mem = DefaultMemThreshold
		log.Debugf("[%s / %s] Memory threshold is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Mem)
	}
	m.SetThreshold(&s.Threshold)
	return nil
}

// SetThreshold set the threshold of the memory
func (m *Mem) SetThreshold(t *Threshold) {
	m.Threshold = t.Mem
}

// Parse a string to a CPU struct
func (m *Mem) Parse(s []string) error {
	if len(s) < m.OutputLines() {
		return fmt.Errorf("invalid memory output")
	}
	mem := strings.Split(s[0], " ")
	if len(mem) < 3 {
		return fmt.Errorf("invalid memory output")
	}
	m.Used = int(strInt(mem[0]))
	m.Total = int(strInt(mem[1]))
	m.Usage = strFloat(mem[2])
	return nil
}

// UsageInfo returns the usage info of the memory
func (m *Mem) UsageInfo() string {
	return fmt.Sprintf("Memory: %.2f%%", m.Usage)
}

// CheckThreshold check the cpu usage
func (m *Mem) CheckThreshold() (bool, string) {
	if m.Threshold > 0 && m.Threshold <= m.Usage/100 {
		return false, "Memory Shortage!"
	}
	return true, ""
}

// ExportMetrics export the cpu metrics
func (m *Mem) ExportMetrics(name string, g *prometheus.GaugeVec) {
	g.With(prometheus.Labels{
		"host":  name,
		"state": "used",
	}).Set(float64(m.Used))

	g.With(prometheus.Labels{
		"host":  name,
		"state": "available",
	}).Set(float64(m.Total - m.Used))

	g.With(prometheus.Labels{
		"host":  name,
		"state": "total",
	}).Set(float64(m.Total))

	g.With(prometheus.Labels{
		"host":  name,
		"state": "usage",
	}).Set(m.Usage)
}
