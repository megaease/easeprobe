package kafka

import (
	"context"

	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of driver
const Kind string = "Kafka"

// Kafka is the Kafka client
type Kafka struct {
	conf.Options `yaml:",inline"`
	ConnStr      string          `yaml:"conn_str"`
	Context      context.Context `yaml:"-"`
}

// New create a Redis client
func New(opt conf.Options) Kafka {
	return Kafka{
		Options: opt,
		ConnStr: "",
		Context: context.Background(),
	}
}

// Kind return the name of client
func (k Kafka) Kind() string {
	return Kind
}

// Probe do the health check
func (k Kafka) Probe() (bool, string) {

	var dialer *kafka.Dialer

	if len(k.Password) > 0 {
		dialer = &kafka.Dialer{
			Timeout: k.Timeout,
			SASLMechanism: plain.Mechanism{
				Username: k.Username,
				Password: k.Password,
			},
		}
	} else {
		dialer = &kafka.Dialer{
			Timeout:       k.Timeout,
			SASLMechanism: nil,
		}
	}

	ctx, cancel := context.WithTimeout(k.Context, k.Timeout)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", k.Host)

	if err != nil {
		return false, err.Error()
	}

	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false, err.Error()
	}

	m := map[string]struct{}{}

	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	for t := range m {
		log.Debugf("[%s] Topic Name - %s", k.Kind(), t)
	}

	return true, "Check Kafka Server Successfully!"

}
