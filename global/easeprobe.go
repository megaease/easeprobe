package global

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// EaseProbe is the information of the program
type EaseProbe struct {
	Name    string `yaml:"name"`
	IconURL string `yaml:"icon"`
	Version string `yaml:"version"`
	Host    string `yaml:"host"`
}

var easeProbe *EaseProbe

// InitEaseProbe the EaseProbe
func InitEaseProbe(name, icon string) {
	host, err := os.Hostname()
	if err != nil {
		log.Errorf("Get Hostname Failed: %s", err)
		host = "unknown"
	}
	easeProbe = &EaseProbe{
		Name:    name,
		IconURL: icon,
		Version: Ver,
		Host:    host,
	}
}

// GetEaseProbe return the EaseProbe
func GetEaseProbe() *EaseProbe {
	return easeProbe
}

// FooterString return the footer string
// e.g. "EaseProbe v1.0.0 @ localhost"
func FooterString() string {
	return easeProbe.Name + " " + easeProbe.Version + " @ " + easeProbe.Host
}
