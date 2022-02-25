package tcp

import (
	"net"
	"time"

	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// Kind is the type
const Kind string = "tcp"

// TCP implements a config for TCP
type TCP struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`

	//Control Option
	Timeout      time.Duration `yaml:"timeout,omitempty"`
	TimeInterval time.Duration `yaml:"interval,omitempty"`
}

// Kind return the HTTP kind
func (t *TCP) Kind() string {
	return Kind
}

// Interval get the interval
func (t *TCP) Interval() time.Duration {
	return t.TimeInterval
}

// Config HTTP Config Object
func (t *TCP) Config() error {

	if t.Timeout <= 0 {
		t.Timeout = time.Second * 30
	}

	if t.TimeInterval <= 0 {
		t.TimeInterval = time.Second * 60
	}

	return nil
}

// Probe return the checking result
func (t *TCP) Probe() probe.Result {

	now := time.Now()
	result := probe.Result{
		Name:          t.Name,
		Endpoint:      t.Host,
		StartTime:     now.Unix(),
		RoundTripTime: probe.ConfigDuration{},
		Status:        "",
		Message:       "",
	}

	conn, err := net.DialTimeout("tcp", t.Host, t.Timeout)
	result.RoundTripTime.Duration = time.Since(now)
	if err != nil {
		log.Errorf("error: %v\n", err)
		result.Status = probe.StatusDown.String()
		return result
	}
	conn.Close()
	result.Status = probe.StatusUp.String()

	return result
}
