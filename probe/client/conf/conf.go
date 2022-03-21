package conf

import (
	"fmt"
	"strings"
	"time"
)

// Driver Interface
type Driver interface {
	Kind() string
	Probe() (bool, string)
}

// DriverType is the client driver
type DriverType int

// The Driver Type of native client
const (
	MySQL DriverType = iota
	Redis
	Kafka
	Mongo
	Unknown
)

// Options implements the configuration for native client
type Options struct {
	Name       string     `yaml:"name"`
	Host       string     `yaml:"host"`
	DriverType DriverType `yaml:"driver"`
	Username   string     `yaml:"username"`
	Password   string     `yaml:"password"`

	//TLS
	CA   string `yaml:"ca,omitempty"`
	Cert string `yaml:"cert,omitempty"`
	Key  string `yaml:"key,omitempty"`

	//Control Option
	Timeout      time.Duration `yaml:"timeout,omitempty"`
	TimeInterval time.Duration `yaml:"interval,omitempty"`
}

// String convert the DriverType to string
func (d DriverType) String() string {
	switch d {
	case MySQL:
		return "MySQL"
	case Redis:
		return "Redis"
	case Kafka:
		return "Kafka"
	case Mongo:
		return "Mongo"
	}
	return "Unknown"
}

// DriverType convert the string to DriverType
func (d *DriverType) DriverType(driver string) DriverType {
	switch driver {
	case "mysql":
		return MySQL
	case "redis":
		return Redis
	case "kafka":
		return Kafka
	case "mongo":
		return Mongo
	}
	return Unknown
}

// UnmarshalYAML is unmarshal the driver type
func (d *DriverType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*d = d.DriverType(strings.ToLower(s))
	return nil
}

// UnmarshalJSON is Unmarshal the driver type
func (d *DriverType) UnmarshalJSON(b []byte) (err error) {
	*d = d.DriverType(strings.ToLower(string(b)))
	return nil
}

// MarshalJSON is marshal the driver
func (d *DriverType) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}
