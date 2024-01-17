/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package global

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/constraints"
)

const (
	// Org is the organization
	Org = "MegaEase"
	// DefaultProg is the program name
	DefaultProg = "EaseProbe"
	// DefaultIconURL is the default icon which used in Slack or Discord
	DefaultIconURL = "https://megaease.com/favicon.png"
)

var (
	// Ver is the program version
	// It will be set by the build script
	// go build -ldflags "-X github.com/megaease/easeprobe/global.Ver=1.0.0"
	Ver = "v1.7.0"
	//OrgProg combine organization and program
	OrgProg = Org + " " + DefaultProg
	//OrgProgVer combine organization and program and version
	OrgProgVer = Org + " " + DefaultProg + "/" + Ver
)

const (
	// DefaultRetryTimes is 3 times
	DefaultRetryTimes = 3
	// DefaultRetryInterval is 5 seconds
	DefaultRetryInterval = time.Second * 5
	// DefaultTimeFormat is "2006-01-02 15:04:05 Z0700"
	DefaultTimeFormat = "2006-01-02 15:04:05 Z0700"
	// DefaultTimeZone is "UTC"
	DefaultTimeZone = "UTC"
	// DefaultProbeInterval is 1 minutes
	DefaultProbeInterval = time.Second * 60
	// DefaultTimeOut is 30 seconds
	DefaultTimeOut = time.Second * 30
	// DefaultChannelName  is the default wide channel name
	DefaultChannelName = "__EaseProbe_Channel__"
	// DefaultStatusChangeThresholdSetting is the threshold of status change
	DefaultStatusChangeThresholdSetting = 1
	// DefaultNotificationStrategy is the default notify strategy
	DefaultNotificationStrategy = RegularStrategy
	// DefaultMaxNotificationTimes is the default max notification times
	DefaultMaxNotificationTimes = 1
	// DefaultNotificationFactor is the default notification factor
	DefaultNotificationFactor = 1
	// DefaultConfigFileCheckInterval is the default config file checking interval
	DefaultConfigFileCheckInterval = time.Second * 5
)

const (
	// DefaultHTTPServerIP is the default ip of the HTTP server
	DefaultHTTPServerIP = "0.0.0.0"
	// DefaultHTTPServerPort is the default port of the HTTP server
	DefaultHTTPServerPort = "8181"
	// DefaultPageSize is the default page size
	DefaultPageSize = 100
	// DefaultAccessLogFile is the default access log file name
	DefaultAccessLogFile = "access.log"
	// DefaultDataFile is the default data file name
	DefaultDataFile = "data/data.yaml"
	// DefaultPIDFile is the default pid file name
	DefaultPIDFile = "easeprobe.pid"
)

const (
	// DefaultMaxLogSize is the default max log size
	DefaultMaxLogSize = 10 // 10M
	// DefaultMaxLogAge is the default max log age
	DefaultMaxLogAge = 7 // 7 days
	// DefaultMaxBackups is the default backup file number
	DefaultMaxBackups = 5 // file
	// DefaultLogCompress is the default compress log
	DefaultLogCompress = true
)

// Retry is the settings of retry
type Retry struct {
	Times    int           `yaml:"times" json:"times,omitempty" jsonschema:"title=Retry Times,description=how many times need to retry,minimum=1"`
	Interval time.Duration `yaml:"interval" json:"interval,omitempty" jsonschema:"type=string,format=duration,title=Retry Interval,description=the interval between each retry"`
}

// TLS is the configuration for TLS files
type TLS struct {
	CA       string `yaml:"ca" json:"ca,omitempty" jsonschema:"title=CA File,description=the CA file path"`
	Cert     string `yaml:"cert" json:"cert,omitempty" jsonschema:"title=Cert File,description=the Cert file path"`
	Key      string `yaml:"key" json:"key,omitempty" jsonschema:"title=Key File,description=the Key file path"`
	Insecure bool   `yaml:"insecure" json:"insecure,omitempty" jsonschema:"title=Insecure,description=whether to skip the TLS verification"`
}

// The normalize() function logic as below:
// - if both global and local are not set, then return the _default.
// - if set the global, but not the local, then return the global
// - if set the local, but not the global, then return the local
// - if both global and local are set, then return the local
func normalize[T constraints.Ordered](global, local, valid, _default T) T {
	// if the val is invalid, then assign the default value
	if local <= valid {
		local = _default
		//if the global configuration is validated, assign the global
		if global > valid {
			local = global
		}
	}
	return local
}

// ReverseMap just reverse the map from [key, value] to [value, key]
func ReverseMap[K comparable, V comparable](m map[K]V) map[V]K {
	n := make(map[V]K, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

// EnumMarshalYaml is a help function to marshal the enum to yaml
func EnumMarshalYaml[T comparable](m map[T]string, v T, typename string) (interface{}, error) {
	if val, ok := m[v]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("%v is not a valid %s", v, typename)
}

// EnumMarshalJSON is a help function to marshal the enum to JSON
func EnumMarshalJSON[T comparable](m map[T]string, v T, typename string) ([]byte, error) {
	if val, ok := m[v]; ok {
		return []byte(fmt.Sprintf(`"%s"`, val)), nil
	}
	return nil, fmt.Errorf("%v is not a valid %s", v, typename)
}

// EnumUnmarshalYaml is a help function to unmarshal the enum from yaml
func EnumUnmarshalYaml[T comparable](unmarshal func(interface{}) error, m map[string]T, v *T, init T, typename string) error {
	var str string
	*v = init
	if err := unmarshal(&str); err != nil {
		return err
	}
	if val, ok := m[strings.ToLower(str)]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("%v is not a valid %s", str, typename)
}

// EnumUnmarshalJSON is a help function to unmarshal the enum from JSON
func EnumUnmarshalJSON[T comparable](b []byte, m map[string]T, v *T, init T, typename string) error {
	var str string
	*v = init
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if val, ok := m[strings.ToLower(str)]; ok {
		*v = val
		return nil
	}
	return fmt.Errorf("%v is not a valid %s", str, typename)
}

// Config return a tls.Config object
func (t *TLS) Config() (*tls.Config, error) {
	if len(t.CA) <= 0 {
		// the insecure is true but no ca/cert/key, then return a tls config
		if t.Insecure == true {
			log.Debug("[TLS] Insecure is true but the CA is empty, return a tls config")
			return &tls.Config{InsecureSkipVerify: true}, nil
		}
		return nil, nil
	}

	cert, err := os.ReadFile(t.CA)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	// only have CA file, go TLS
	if len(t.Cert) <= 0 || len(t.Key) <= 0 {
		log.Debug("[TLS] Only have CA file, go TLS")
		return &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: t.Insecure,
		}, nil
	}

	// have both CA and cert/key, go mTLS way
	log.Debug("[TLS] Have both CA and cert/key, go mTLS way")
	certificate, err := tls.LoadX509KeyPair(t.Cert, t.Key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: t.Insecure,
	}, nil
}

// ErrNoRetry is the error need not retry
type ErrNoRetry struct {
	Message string
}

func (e *ErrNoRetry) Error() string {
	return e.Message
}

// DoRetry is a help function to retry the function if it returns error
func DoRetry(kind, name, tag string, r Retry, fn func() error) error {
	var err error
	for i := 0; i < r.Times; i++ {
		err = fn()
		_, ok := err.(*ErrNoRetry)
		if err == nil || ok {
			return err
		}
		log.Warnf("[%s / %s / %s] Retried to send %d/%d - %v", kind, name, tag, i+1, r.Times, err)

		// last time no need to sleep
		if i < r.Times-1 {
			time.Sleep(r.Interval)
		}
	}
	return fmt.Errorf("[%s / %s / %s] failed after %d retries - %v", kind, name, tag, r.Times, err)
}

// GetWorkDir return the current working directory
func GetWorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Warnf("Cannot get the current directory: %v, using $HOME directory!", err)
		dir, err = os.UserHomeDir()
		if err != nil {
			log.Warnf("Cannot get the user home directory: %v, using /tmp directory!", err)
			dir = os.TempDir()
		}
	}
	return dir
}

// MakeDirectory return the writeable filename
func MakeDirectory(filename string) string {
	dir, file := filepath.Split(filename)
	if len(dir) <= 0 {
		dir = GetWorkDir()
	}
	if len(file) <= 0 {
		return dir
	}
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Warnf("Cannot get the user home directory: %v, using /tmp directory as home", err)
			home = os.TempDir()
		}
		dir = filepath.Join(home, dir[2:])
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Warnf("Cannot get the absolute path: %v", err)
		dir = GetWorkDir()
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Warnf("Cannot create the directory: %v", err)
			dir = GetWorkDir()
		}
	}

	return filepath.Join(dir, file)
}

// CommandLine will return the whole command line which includes command and all arguments
func CommandLine(cmd string, args []string) string {
	result := cmd
	for _, arg := range args {
		result += " " + arg
	}
	return result
}

// EscapeQuote escape the string the single quote, double quote, and backtick
func EscapeQuote(str string) string {
	type Escape struct {
		From string
		To   string
	}
	escape := []Escape{
		{From: "`", To: ""}, // remove the backtick
		{From: `\`, To: `\\`},
		{From: `'`, To: `\'`},
		{From: `"`, To: `\"`},
	}

	for _, e := range escape {
		str = strings.ReplaceAll(str, e.From, e.To)
	}
	return str
}
