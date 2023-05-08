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

// Package conf is the configuration of the application
package conf

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	httpClient "net/http"
	netUrl "net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/megaease/easeprobe/channel"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/client"
	"github.com/megaease/easeprobe/probe/host"
	"github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/ping"
	"github.com/megaease/easeprobe/probe/shell"
	"github.com/megaease/easeprobe/probe/ssh"
	"github.com/megaease/easeprobe/probe/tcp"
	"github.com/megaease/easeprobe/probe/tls"

	"github.com/invopop/jsonschema"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var config *Conf

// Get return the global configuration
func Get() *Conf {
	return config
}

// Schedule is the schedule.
type Schedule int

// Schedule enum
const (
	None Schedule = iota
	Minutely
	Hourly
	Daily
	Weekly
	Monthly
)

var scheduleToString = map[Schedule]string{
	Minutely: "minutely",
	Hourly:   "hourly",
	Daily:    "daily",
	Weekly:   "weekly",
	Monthly:  "monthly",
	None:     "none",
}

var stringToSchedule = global.ReverseMap(scheduleToString)

// MarshalYAML marshal the configuration to yaml
func (s Schedule) MarshalYAML() (interface{}, error) {
	return global.EnumMarshalYaml(scheduleToString, s, "Schedule")
}

// UnmarshalYAML is unmarshal the debug level
func (s *Schedule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return global.EnumUnmarshalYaml(unmarshal, stringToSchedule, s, None, "Schedule")
}

// Notify is the settings of notification
type Notify struct {
	Retry global.Retry `yaml:"retry" json:"retry,omitempty" jsonschema:"title=retry,description=the retry settings"`
	Dry   bool         `yaml:"dry" json:"dry,omitempty" jsonschema:"title=dry,description=set true to make the notification dry run and will not be sent the message,default=false"`
}

// Probe is the settings of prober
type Probe struct {
	Interval                             time.Duration `yaml:"interval" json:"interval,omitempty" jsonschema:"type=string,format=duration,title=Probe Interval,description=the interval of probe,default=1m"`
	Timeout                              time.Duration `yaml:"timeout" json:"timeout,omitempty" jsonschema:"type=string,format=duration,title=Probe Timeout,description=the timeout of probe,default=30s"`
	global.StatusChangeThresholdSettings `yaml:",inline" json:",inline"`
	global.NotificationStrategySettings  `yaml:"alert" json:"alert" jsonschema:"title=Alert,description=the alert settings"`
}

// SLAReport is the settings for SLA report
type SLAReport struct {
	Schedule Schedule `yaml:"schedule" json:"schedule" jsonschema:"type=string,enum=none,enum=minutely,enum=hourly,enum=daily,enum=weekly,enum=monthly,title=Schedule,description=the schedule of SLA report"`
	Time     string   `yaml:"time" json:"time,omitempty" jsonschema:"format=time,title=Time,description=the time of SLA report need to send out,example=23:59:59+08:00"`
	//Debug    bool     `yaml:"debug" json:"debug,omitempty" jsonschema:"title=Debug,description=if true the SLA report will be printed to stdout,default=false"`
	DataFile string   `yaml:"data" json:"data,omitempty" jsonschema:"title=Data File,description=the data file of SLA report, absolute path. ('-' means no SLA persistent data)"`
	Backups  int      `yaml:"backups" json:"backups,omitempty" jsonschema:"title=Backups,description=the number of backups of SLA report,default=5"`
	Channels []string `yaml:"channels" json:"channels,omitempty" jsonschema:"title=Channels,description=the channels of SLA report"`
}

// HTTPServer is the settings of http server
type HTTPServer struct {
	IP              string        `yaml:"ip" json:"ip" jsonschema:"title=Web Server IP,description=the local ip address of the http server need to listen on,example=0.0.0.0"`
	Port            string        `yaml:"port" json:"port" jsonschema:"type=integer,title=Web Server Port,description=port of the http server,default=8181"`
	AutoRefreshTime time.Duration `yaml:"refresh" json:"refresh,omitempty" jsonschema:"type=string,title=Auto Refresh Time,description=auto refresh time of the http server,example=5s"`
	AccessLog       Log           `yaml:"log" json:"log,omitempty" jsonschema:"title=Access Log,description=access log of the http server"`
}

// Settings is the EaseProbe configuration
type Settings struct {
	Name       string     `yaml:"name" json:"name,omitempty" jsonschema:"title=EaseProbe Name,description=The name of the EaseProbe instance,default=EaseProbe"`
	IconURL    string     `yaml:"icon" json:"icon,omitempty" jsonschema:"title=Icon URL,description=The URL of the icon of the EaseProbe instance"`
	PIDFile    string     `yaml:"pid" json:"pid,omitempty" jsonschema:"title=PID File,description=The PID file of the EaseProbe instance ('' or '-' means no PID file)"`
	Log        Log        `yaml:"log" json:"log,omitempty" jsonschema:"title=EaseProbe Log,description=The log settings of the EaseProbe instance"`
	TimeFormat string     `yaml:"timeformat" json:"timeformat,omitempty" jsonschema:"title=Time Format,description=The time format of the EaseProbe instance,default=2006-01-02 15:04:05Z07:00"`
	TimeZone   string     `yaml:"timezone" json:"timezone,omitempty" jsonschema:"title=Time Zone,description=The time zone of the EaseProbe instance,example=Asia/Shanghai,example=Europe/Berlin,default=UTC"`
	Probe      Probe      `yaml:"probe" json:"probe,omitempty" jsonschema:"title=Probe Settings,description=The global probe settings of the EaseProbe instance"`
	Notify     Notify     `yaml:"notify" json:"notify,omitempty" jsonschema:"title=Notify Settings,description=The global notify settings of the EaseProbe instance"`
	SLAReport  SLAReport  `yaml:"sla" json:"sla,omitempty" jsonschema:"title=SLA Report Settings,description=The SLA report settings of the EaseProbe instance"`
	HTTPServer HTTPServer `yaml:"http" json:"http,omitempty" jsonschema:"title=HTTP Server Settings,description=The HTTP server settings of the EaseProbe instance"`
}

// Conf is Probe configuration
type Conf struct {
	Version  string          `yaml:"version" json:"version,omitempty" jsonschema:"title=Version,description=Version of the EaseProbe configuration"`
	HTTP     []http.HTTP     `yaml:"http" json:"http,omitempty" jsonschema:"title=HTTP Probe,description=HTTP Probe Configuration"`
	TCP      []tcp.TCP       `yaml:"tcp" json:"tcp,omitempty" jsonschema:"title=TCP Probe,description=TCP Probe Configuration"`
	Shell    []shell.Shell   `yaml:"shell" json:"shell,omitempty" jsonschema:"title=Shell Probe,description=Shell Probe Configuration"`
	Client   []client.Client `yaml:"client" json:"client,omitempty" jsonschema:"title=Native Client Probe,description=Native Client Probe Configuration"`
	SSH      ssh.SSH         `yaml:"ssh" json:"ssh,omitempty" jsonschema:"title=SSH Probe,description=SSH Probe Configuration"`
	TLS      []tls.TLS       `yaml:"tls" json:"tls,omitempty" jsonschema:"title=TLS Probe,description=TLS Probe Configuration"`
	Host     host.Host       `yaml:"host" json:"host,omitempty" jsonschema:"title=Host Probe,description=Host Probe Configuration"`
	Ping     []ping.Ping     `yaml:"ping" json:"ping,omitempty" jsonschema:"title=Ping Probe,description=Ping Probe Configuration"`
	Notify   notify.Config   `yaml:"notify" json:"notify,omitempty" jsonschema:"title=Notification,description=Notification Configuration"`
	Settings Settings        `yaml:"settings" json:"settings,omitempty" jsonschema:"title=Global Settings,description=EaseProbe Global configuration"`
}

// JSONSchema return the json schema of the configuration
func JSONSchema() (string, error) {
	r := new(jsonschema.Reflector)

	// The Struct name could be same, but the package name is different
	// For example, all of the notification plugins have the same struct name - `NotifyConfig`
	// This would cause the json schema to be wrong `$ref` to the same name.
	// the following code is to fix this issue by adding the package name to the struct name
	// p.s. this issue has been reported in: https://github.com/invopop/jsonschema/issues/42
	r.Namer = func(t reflect.Type) string {
		name := t.Name()
		if t.Kind() == reflect.Struct {
			v := reflect.New(t)
			vt := v.Elem().Type()
			name = vt.PkgPath() + "/" + vt.Name()
			name = strings.TrimPrefix(name, "github.com/megaease/easeprobe/")
			name = strings.ReplaceAll(name, "/", "_")
			log.Debugf("The struct name has been replaced [%s ==> %s]", t.Name(), name)
		}
		return name
	}

	schema := r.Reflect(&Conf{})

	resBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(resBytes), nil
}

// Check if string is a url
func isExternalURL(url string) bool {
	if _, err := netUrl.ParseRequestURI(url); err != nil {
		log.Debugf("ParseRequestedURI: %s failed to parse with error %v", url, err)
		return false
	}

	parts, err := netUrl.Parse(url)
	if err != nil || parts.Host == "" || !strings.HasPrefix(parts.Scheme, "http") {
		log.Debugf("Parse: %s failed Scheme: %s, Host: %s (err: %v)", url, parts.Scheme, parts.Host, err)
		return false
	}

	return true
}

func getYamlFileFromHTTP(url string) ([]byte, error) {
	r, err := httpClient.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if os.Getenv("HTTP_AUTHORIZATION") != "" {
		r.Header.Set("Authorization", os.Getenv("HTTP_AUTHORIZATION"))
	}

	httpClientObject := httpClient.Client{}
	if os.Getenv("HTTP_TIMEOUT") != "" {
		timeout, err := strconv.ParseInt(os.Getenv("HTTP_TIMEOUT"), 10, 64)
		if err != nil {
			return nil, err
		}
		httpClientObject.Timeout = time.Duration(timeout) * time.Second
	}

	resp, err := httpClientObject.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getYamlFileFromFile(path string) ([]byte, error) {
	f, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}
	if f.IsDir() {
		return mergeYamlFiles(path)
	}
	return ioutil.ReadFile(path)
}

func getYamlFile(path string) ([]byte, error) {
	if isExternalURL(path) {
		return getYamlFileFromHTTP(path)
	}
	return getYamlFileFromFile(path)
}

// previousYAMLFile is the content of the configuration file
var previousYAMLFile []byte

// ResetPreviousYAMLFile resets the previousYAMLFile
func ResetPreviousYAMLFile() {
	previousYAMLFile = nil
}

// IsConfigModified checks if the configuration file is modified
func IsConfigModified(path string) bool {

	var content []byte
	var err error
	if isExternalURL(path) {
		content, err = getYamlFileFromHTTP(path)
	} else {
		content, err = getYamlFileFromFile(path)
	}

	if err != nil {
		log.Warnf("Failed to get the configuration file [%s]: %v", path, err)
		return false
	}

	// if it is the fisrt time to read the configuration file, we will not restart the program
	if previousYAMLFile == nil {
		previousYAMLFile = content
		return false
	}

	//  if the configuration file is invalid, we will not restart the program
	testConf := Conf{}
	err = yaml.Unmarshal(content, &testConf)
	if err != nil {
		log.Warnf("Invalid configuration file [%s]: %v", path, err)
		return false
	}

	// check if the configuration file is modified
	modified := !bytes.Equal(content, previousYAMLFile)
	previousYAMLFile = content
	return modified
}

// New read the configuration from yaml
func New(conf *string) (*Conf, error) {
	c := Conf{
		HTTP:   []http.HTTP{},
		TCP:    []tcp.TCP{},
		Shell:  []shell.Shell{},
		Client: []client.Client{},
		SSH: ssh.SSH{
			Bastion: &ssh.BastionMap,
			Servers: []ssh.Server{},
		},
		TLS: []tls.TLS{},
		Host: host.Host{
			Bastion: &host.BastionMap,
			Servers: []host.Server{},
		},
		Notify: notify.Config{},
		Settings: Settings{
			Name:       global.DefaultProg,
			IconURL:    global.DefaultIconURL,
			PIDFile:    filepath.Join(global.GetWorkDir(), global.DefaultPIDFile),
			Log:        NewLog(),
			TimeFormat: "2006-01-02 15:04:05 UTC",
			TimeZone:   "UTC",
			Probe: Probe{
				Interval: global.DefaultProbeInterval,
				Timeout:  global.DefaultTimeOut,
			},
			Notify: Notify{
				Retry: global.Retry{
					Times:    global.DefaultRetryTimes,
					Interval: global.DefaultRetryInterval,
				},
				Dry: false,
			},
			SLAReport: SLAReport{
				Schedule: Daily,
				Time:     "00:00",
				//Debug:    false,
				DataFile: global.DefaultDataFile,
				Backups:  global.DefaultMaxBackups,
				Channels: []string{global.DefaultChannelName},
			},
			HTTPServer: HTTPServer{
				IP:        global.DefaultHTTPServerIP,
				Port:      global.DefaultHTTPServerPort,
				AccessLog: NewLog(),
			},
		},
	}
	y, err := getYamlFile(*conf)
	if err != nil {
		log.Errorf("error: %v ", err)
		return &c, err
	}

	y = []byte(os.ExpandEnv(string(y)))

	err = yaml.Unmarshal(y, &c)
	if err != nil {
		log.Errorf("error: %v", err)
		return &c, err
	}

	// Initialization
	c.Settings.Log.InitLog(nil)
	global.InitEaseProbeWithTime(c.Settings.Name, c.Settings.IconURL,
		c.Settings.TimeFormat, c.Settings.TimeZone)
	c.initData()

	ssh.BastionMap.ParseAllBastionHost()
	host.BastionMap.ParseAllBastionHost()

	// pass the dry run to the channel
	channel.SetDryNotify(c.Settings.Notify.Dry)

	config = &c

	log.Infoln("Load the configuration file successfully!")
	if log.GetLevel() >= log.DebugLevel {
		s, err := yaml.Marshal(c)
		if err != nil {
			log.Debugf("%v\n%+v", err, c)
		} else {
			log.Debugf("\n%s", string(s))
		}
	}

	return &c, err
}

// InitAllLogs initialize all logs
func (conf *Conf) InitAllLogs() {

	conf.Settings.Log.InitLog(nil)
	conf.Settings.Log.LogInfo("Application")

	conf.Settings.HTTPServer.AccessLog.InitLog(log.New())
	conf.Settings.HTTPServer.AccessLog.LogInfo("Web Access")
}

func (conf *Conf) initData() {

	// Check if we are explicitly disabled
	if strings.TrimSpace(conf.Settings.SLAReport.DataFile) == "-" {
		log.Infof("SLA data disabled by configuration. Skipping SLA data store...")
		return
	}

	// Check if we are empty and use global.DefaultDataFile
	if strings.TrimSpace(conf.Settings.SLAReport.DataFile) == "" {
		conf.Settings.SLAReport.DataFile = global.DefaultDataFile
	}

	dir, _ := filepath.Split(conf.Settings.SLAReport.DataFile)
	// if dir part is not empty
	if strings.TrimSpace(dir) != "" {
		// check for `dir`` existence and create intermediate folders
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Infof("Creating base directory for data file!")
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Warnf("Failed to create base directory for data file: %s", err.Error())
				return
			}
		}
	}

	// check if the data file exists and is a regular file
	dataInfo, err := os.Stat(conf.Settings.SLAReport.DataFile)
	if os.IsNotExist(err) || !dataInfo.Mode().IsRegular() {
		log.Infof("The data file %s, was not found!", conf.Settings.SLAReport.DataFile)
		return
	}

	if err := probe.LoadDataFromFile(conf.Settings.SLAReport.DataFile); err != nil {
		log.Warnf("Cannot load data from file(%s): %v", conf.Settings.SLAReport.DataFile, err)
	}

	probe.CleanDataFile(conf.Settings.SLAReport.DataFile, conf.Settings.SLAReport.Backups)
}

// isProbe checks whether a interface is a probe type
func isProbe(t reflect.Type) bool {
	modelType := reflect.TypeOf((*probe.Prober)(nil)).Elem()
	return t.Implements(modelType)
}

// AllProbers return all probers
func (conf *Conf) AllProbers() []probe.Prober {
	log.Debugf("--------- Process the probers settings ---------")
	return allProbersHelper(*conf)
}

func allProbersHelper(i interface{}) []probe.Prober {

	var probers []probe.Prober
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	if t.Kind() != reflect.Struct {
		return probers
	}

	for i := 0; i < t.NumField(); i++ {
		tField := t.Field(i).Type.Kind()
		if tField == reflect.Struct {
			probers = append(probers, allProbersHelper(v.Field(i).Interface())...)
			continue
		}
		if tField != reflect.Slice {
			continue
		}

		vField := v.Field(i)
		for j := 0; j < vField.Len(); j++ {
			if !isProbe(vField.Index(j).Addr().Type()) {
				//log.Debugf("%s is not a probe type", vField.Index(j).Type())
				continue
			}

			log.Debugf("--> %s / %s / %+v", t.Field(i).Name, t.Field(i).Type.Kind(), vField.Index(j))
			probers = append(probers, vField.Index(j).Addr().Interface().(probe.Prober))
		}
	}

	return probers
}

// isNotify checks whether a interface is a Notify type
func isNotify(t reflect.Type) bool {
	modelType := reflect.TypeOf((*notify.Notify)(nil)).Elem()
	return t.Implements(modelType)
}

// AllNotifiers return all notifiers
func (conf *Conf) AllNotifiers() []notify.Notify {
	var notifies []notify.Notify

	log.Debugf("--------- Process the notification settings ---------")
	t := reflect.TypeOf(conf.Notify)
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() != reflect.Slice {
			continue
		}
		v := reflect.ValueOf(conf.Notify).Field(i)
		for j := 0; j < v.Len(); j++ {
			if !isNotify(v.Index(j).Addr().Type()) {
				log.Debugf("%s is not a notify type", v.Index(j).Type())
				continue
			}
			log.Debugf("--> %s - %s - %+v", t.Field(i).Name, t.Field(i).Type.Kind(), v.Index(j))
			notifies = append(notifies, v.Index(j).Addr().Interface().(notify.Notify))
		}
	}

	return notifies
}
