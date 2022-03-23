package global

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"
)

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

// TLS is the configuration for TLS files
type TLS struct {
	CA   string `yaml:"ca"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

func normalizeTimeDuration(global, local, valid, _default time.Duration) time.Duration {
	// if the val is in valid, the assign the default value
	if local <= valid {
		local = _default
		//if the global configuration is validated, assign the global
		if global > valid {
			local = global
		}
	}
	return local
}

func normalizeInteger(global, local, valid, _default int) int {
	// if the val is in valid, the assign the default value
	if local <= valid {
		local = _default
		//if the global configuration is validated, assign the global
		if global > valid {
			local = global
		}
	}
	return local
}

// Config return a tls.Config object
func (t *TLS) Config() (*tls.Config, error) {
	if len(t.CA) <= 0 || len(t.Cert) <= 0 || len(t.Key) <= 0 {
		return nil, nil
	}

	cert, err := ioutil.ReadFile(t.CA)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	certificate, err := tls.LoadX509KeyPair(t.Cert, t.Key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{certificate},
	}, nil
}
