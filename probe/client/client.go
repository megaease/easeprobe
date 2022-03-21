package client

import (
	"fmt"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/megaease/easeprobe/probe/client/kafka"
	"github.com/megaease/easeprobe/probe/client/mongo"
	"github.com/megaease/easeprobe/probe/client/mysql"
	"github.com/megaease/easeprobe/probe/client/redis"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of probe
const Kind string = "client"

// Client implements the structure of client
type Client struct {
	//Embed structure
	conf.Options `yaml:",inline"`

	result *probe.Result `yaml:"-"`
	client conf.Driver   `yaml:"-"`
}

// Kind return the Client kind
func (c *Client) Kind() string {
	return Kind
}

// Interval get the interval
func (c *Client) Interval() time.Duration {
	return c.TimeInterval
}

// Result get the probe result
func (c *Client) Result() *probe.Result {
	return c.result
}

// Config Client Config Object
func (c *Client) Config(gConf global.ProbeSettings) error {

	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.TimeInterval = gConf.NormalizeInterval(c.TimeInterval)

	c.configClientDriver()

	c.result = probe.NewResult()
	c.result.Name = c.Name
	c.result.Endpoint = c.Host
	c.result.PreStatus = probe.StatusInit
	c.result.TimeFormat = gConf.TimeFormat

	log.Debugf("[%s] configuration: %+v, %+v", c.Kind(), c, c.Result())
	return nil
}

func (c *Client) configClientDriver() {
	switch c.DriverType {
	case conf.MySQL:
		c.client = mysql.New(c.Options)
	case conf.Redis:
		c.client = redis.New(c.Options)
	case conf.Mongo:
		c.client = mongo.New(c.Options)
	case conf.Kafka:
		c.client = kafka.New(c.Options)
	default:
		c.DriverType = conf.Unknown
	}

}

// Probe return the checking result
func (c *Client) Probe() probe.Result {
	if c.DriverType == conf.Unknown {
		c.result.PreStatus = probe.StatusUnknown
		c.result.Status = probe.StatusUnknown
		return *c.result
	}

	now := time.Now()
	c.result.StartTime = now
	c.result.StartTimestamp = now.UnixMilli()

	stat, msg := c.client.Probe()

	c.result.RoundTripTime.Duration = time.Since(now)

	status := probe.StatusUp
	c.result.Message = fmt.Sprintf("%s client checked up successfully!", c.DriverType.String())

	if stat != true {
		c.result.Message = fmt.Sprintf("Error (%s): %s", c.DriverType.String(), msg)
		log.Errorf("[%s / %s / %s] - %s", c.Kind(), c.client.Kind(), c.Name, msg)
		status = probe.StatusDown
	} else {
		log.Debugf("[%s / %s / %s] - %s", c.Kind(), c.client.Kind(), c.Name, msg)
	}

	c.result.PreStatus = c.result.Status
	c.result.Status = status

	c.result.DoStat(c.TimeInterval)
	return *c.result
}
