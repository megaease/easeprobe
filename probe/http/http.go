package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of probe
const Kind string = "http"

// HTTP implements a config for HTTP.
type HTTP struct {
	Name            string            `yaml:"name"`
	URL             string            `yaml:"url"`
	ContentEncoding string            `yaml:"content_encoding,omitempty"`
	Method          string            `yaml:"method,omitempty"`
	Headers         map[string]string `yaml:"headers,omitempty"`
	Body            string            `yaml:"body,omitempty"`

	//Option - HTTP Basic Auth Credentials
	User string `yaml:"username,omitempty"`
	Pass string `yaml:"password,omitempty"`

	//Option - TLS Config
	CA   string `yaml:"ca,omitempty"`
	Cert string `yaml:"cert,omitempty"`
	Key  string `yaml:"key,omitempty"`

	//Control Options
	Timeout      time.Duration `yaml:"timeout,omitempty"`
	TimeInterval time.Duration `yaml:"interval,omitempty"`

	client *http.Client `yaml:"-"`
}

func checkHTTPMethod(m string) bool {

	methods := [...]string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
	for _, method := range methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// Kind return the HTTP kind
func (h *HTTP) Kind() string {
	return Kind
}

// Interval get the interval
func (h *HTTP) Interval() time.Duration {
	return h.TimeInterval
}

// Config HTTP Config Object
func (h *HTTP) Config() error {

	if h.Timeout <= 0 {
		h.Timeout = time.Second * 30
	}

	if h.TimeInterval <= 0 {
		h.TimeInterval = time.Second * 60
	}

	h.client = &http.Client{
		Timeout: h.Timeout,
	}
	if !checkHTTPMethod(h.Method) {
		h.Method = "GET"
	}

	if len(h.CA) > 0 {
		cert, err := ioutil.ReadFile("./certs/ca.crt")
		if err != nil {
			log.Errorf("could not open certificate file: %v\n", err)
			return err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(cert)

		log.Info("Load key pairs - ", h.Cert, h.Key)
		certificate, err := tls.LoadX509KeyPair(h.Cert, h.Key)
		if err != nil {
			log.Errorf("could not load certificate: %v\n", err)
			return err
		}
		h.client = &http.Client{
			Timeout: time.Minute * 3,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      caCertPool,
					Certificates: []tls.Certificate{certificate},
				},
			},
		}

	}
	return nil
}

// Probe return the checking result
func (h *HTTP) Probe() probe.Result {

	req, _ := http.NewRequest(h.Method, h.URL, bytes.NewBuffer([]byte(h.Body)))
	if len(h.User) > 0 && len(h.Pass) > 0 {
		req.SetBasicAuth(h.User, h.Pass)
	}
	if len(h.ContentEncoding) > 0 {
		req.Header.Set("Content-Type", h.ContentEncoding)
	}
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	now := time.Now()
	result := probe.Result{
		Name:          h.Name,
		Endpoint:      h.URL,
		StartTime:     now.Unix(),
		RoundTripTime: probe.ConfigDuration{},
		Status:        "",
		Message:       "",
	}

	resp, err := h.client.Do(req)
	result.RoundTripTime.Duration = time.Since(now)
	if err != nil {
		log.Errorf("error making get request: %v", err)
		result.Status = probe.StatusUnknown.String()
		return result
	}

	if resp.StatusCode >= 500 {
		result.Status = probe.StatusDown.String()
	} else {
		result.Status = probe.StatusUp.String()
	}

	// Read the response body
	defer resp.Body.Close()

	return result
}
