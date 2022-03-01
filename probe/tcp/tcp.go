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

	result *probe.Result `yaml:"-"`
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

	t.result = probe.NewResult()
	t.result.Endpoint = t.Host
	t.result.Name = t.Name
	t.result.PreStatus = probe.StatusInit

	return nil
}

// Probe return the checking result
func (t *TCP) Probe() probe.Result {

	now := time.Now()
	t.result.StartTime = now.UnixMilli()

	conn, err := net.DialTimeout("tcp", t.Host, t.Timeout)
	t.result.RoundTripTime.Duration = time.Since(now)
	status := probe.StatusUp
	if err != nil {
		log.Errorf("error: %v\n", err)
		status = probe.StatusDown
	}
	conn.Close()
	status = probe.StatusUp

	if t.result.PreStatus != probe.StatusInit {
		t.result.PreStatus = t.result.Status
	}
	t.result.Status = status

	return *t.result
}
