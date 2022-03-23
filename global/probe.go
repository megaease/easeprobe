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
	if t <= 0 {
		t = DefaultTimeOut
		if p.Timeout > 0 {
			t = p.Timeout
		}
	}
	return t
}

// NormalizeInterval return a normalized time interval value
func (p *ProbeSettings) NormalizeInterval(t time.Duration) time.Duration {
	if t <= 0 {
		t = DefaultProbeInterval
		if p.Interval > 0 {
			t = p.Interval
		}
	}
	return t
}
