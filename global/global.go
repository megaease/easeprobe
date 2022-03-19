package global

import "time"

const (
	// Org is the organization
	Org = "MegaEase"
	// Prog is the program name
	Prog = "EaseProbe"
	// Ver is the program version
	Ver = "0.1"

	//OrgProg combine organization and program
	OrgProg = Org + " " + Prog
	//OrgProgVer combine organization and program and version
	OrgProgVer = Org + " " + Prog + "/" + Ver
)

const (
	// DefaultRetryTimes is 3 times
	DefaultRetryTimes = 3
	// DefaultRetryInterval is 5 seconds
	DefaultRetryInterval = time.Second * 5
	// DefaultTimeFormat is "2006-01-02 15:04:05 UTC"
	DefaultTimeFormat = "2006-01-02 15:04:05 UTC"
	// DefaultProbeInterval is 1 minutes
	DefaultProbeInterval = time.Second * 60
	// DefaultTimeOut is 30 seconds
	DefaultTimeOut = time.Second * 30
)

// Retry is the settings of retry
type Retry struct {
	Times    int           `yaml:"times"`
	Interval time.Duration `yaml:"interval"`
}

// ProbeSettings is the global probe setting
type ProbeSettings struct {
	TimeFormat string
	Interval   time.Duration
	Timeout    time.Duration
}

// NormalizeTimeOut return a normalized time out value
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

// NotifySettings is the global notification setting
type NotifySettings struct {
	TimeFormat string
	Retry      Retry
}
