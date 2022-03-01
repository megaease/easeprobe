package probe

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// Status is the status of Probe
type Status int

// The status of a probe
const (
	StatusUp Status = iota
	StatusDown
	StatusUnknown
	StatusInit
)

func (s Status) String() string {
	switch s {
	case StatusUp:
		return "up"
	case StatusDown:
		return "down"
	case StatusUnknown:
		return "unknown"
	case StatusInit:
		return "init"
	}
	return "unknown"
}

//ConfigDuration is the struct used for custom the time formation
type ConfigDuration struct {
	time.Duration
}

// UnmarshalJSON is Unmarshal the time
func (d *ConfigDuration) UnmarshalJSON(b []byte) (err error) {
	d.Duration, err = time.ParseDuration(strings.Trim(string(b), `"`))
	return
}

// MarshalJSON is marshal the time
func (d ConfigDuration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// Result is the status of health check
type Result struct {
	Name          string         `json:"name,omitempty"`
	Endpoint      string         `json:"endpoint,omitempty"`
	StartTime     int64          `json:"timestamp,omitempty"`
	RoundTripTime ConfigDuration `json:"rtt,omitempty"`
	Status        Status         `json:"status,omitempty"`
	PreStatus     Status         `json:"prestatus,omitempty"`
	Message       string         `json:"message,omitempty"`
}

// NewResult return a Result object
func NewResult() *Result {
	return &Result{
		Name:          "",
		Endpoint:      "",
		StartTime:     0,
		RoundTripTime: ConfigDuration{0},
		Status:        0,
		PreStatus:     StatusInit,
		Message:       "",
	}
}

func (r Result) String() string {
	j, err := json.Marshal(&r)
	if err != nil {
		log.Printf("error: %v\n", err)
		return ""
	}
	return string(j)
}

// Prober Interface
type Prober interface {
	Kind() string
	Config() error
	Probe() Result
	Interval() time.Duration
}
