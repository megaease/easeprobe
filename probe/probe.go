package probe

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// Prober Interface
type Prober interface {
	Kind() string
	Config() error
	Probe() Result
	Interval() time.Duration
}

// Status is the status of Probe
type Status int

// The status of a probe
const (
	StatusUp Status = iota
	StatusDown
	StatusUnknown
	StatusInit
)

// String convert the Status to string
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

//Status convert the string to Status
func (s *Status) Status(status string) Status {
	switch status {
	case "up":
		return StatusUp
	case "down":
		return StatusDown
	case "unknown":
		return StatusUnknown
	case "init":
		return StatusInit
	}
	return StatusUnknown
}

// UnmarshalJSON is Unmarshal the status
func (s *Status) UnmarshalJSON(b []byte) (err error) {
	*s = s.Status(string(b))
	return nil
}

// MarshalJSON is marshal the status
func (s *Status) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
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
func (d *ConfigDuration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Round(time.Millisecond))), nil
}

// Result is the status of health check
type Result struct {
	Name           string         `json:"name"`
	Endpoint       string         `json:"endpoint"`
	StartTime      time.Time      `json:"time"`
	StartTimestamp int64          `json:"timestamp"`
	RoundTripTime  ConfigDuration `json:"rtt"`
	Status         Status         `json:"status"`
	PreStatus      Status         `json:"prestatus"`
	Message        string         `json:"message"`
}

// NewResult return a Result object
func NewResult() *Result {
	return &Result{
		Name:          "",
		Endpoint:      "",
		StartTime:     time.Now(),
		RoundTripTime: ConfigDuration{0},
		Status:        StatusInit,
		PreStatus:     StatusInit,
		Message:       "",
	}
}

// String convert the object to JSON
func (r *Result) String() string {
	j, err := json.Marshal(&r)
	if err != nil {
		log.Printf("error: %v\n", err)
		return ""
	}
	return string(j)
}

// StringIndent convert the object to indent JSON
func (r *Result) StringIndent() string {
	j, err := json.MarshalIndent(&r, "", "    ")
	if err != nil {
		log.Printf("error: %v\n", err)
		return ""
	}
	return string(j)
}


