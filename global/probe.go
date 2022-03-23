package global

import "time"

// ProbeSettings is the global probe setting
type ProbeSettings struct {
	TimeFormat string
	Interval   time.Duration
	Timeout    time.Duration
}

// NormalizeTimeOut return a normalized timeout value
func (p *ProbeSettings) NormalizeTimeOut(t time.Duration) time.Duration {
	return normalizeTimeDuration(p.Timeout, t, 0, DefaultTimeOut)
}

// NormalizeInterval return a normalized time interval value
func (p *ProbeSettings) NormalizeInterval(t time.Duration) time.Duration {
	return normalizeTimeDuration(p.Interval, t, 0, DefaultProbeInterval)
}
