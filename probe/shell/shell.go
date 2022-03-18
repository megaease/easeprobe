package shell

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of probe
const Kind string = "shell"

// Shell implements a config  for shell command (os.Exec)
type Shell struct {
	Name       string   `yaml:"name"`
	Command    string   `yaml:"cmd"`
	Args       []string `yaml:"args,omitempty"`
	Contain    string   `yaml:"contain,omitempty"`
	NotContain string   `yaml:"not_contain,omitempty"`

	//Control Options
	Timeout      time.Duration `yaml:"timeout,omitempty"`
	TimeInterval time.Duration `yaml:"interval,omitempty"`

	result *probe.Result `yaml:"-"`
}

// Kind return the Shell kind
func (s *Shell) Kind() string {
	return Kind
}

// Interval get the interval
func (s *Shell) Interval() time.Duration {
	return s.TimeInterval
}

// Result get the probe result
func (s *Shell) Result() *probe.Result {
	return s.result
}

// Config Shell Config Object
func (s *Shell) Config(gConf global.ProbeSettings) error {

	if s.Timeout <= 0 {
		s.Timeout = global.DefaultTimeOut
		if gConf.Timeout > 0 {
			s.Timeout = gConf.Timeout
		}
	}

	if s.TimeInterval <= 0 {
		s.TimeInterval = global.DefaultProbeInterval
		if gConf.Interval > 0 {
			s.TimeInterval = gConf.Interval
		}
	}

	s.result = probe.NewResult()
	s.result.Name = s.Name
	s.result.Endpoint = s.CommandLine()
	s.result.PreStatus = probe.StatusInit
	s.result.TimeFormat = gConf.TimeFormat

	log.Debugf("%s configuration: %+v, %+v", s.Kind(), s, s.Result())
	return nil
}

// Probe return the checking result
func (s *Shell) Probe() probe.Result {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	s.result.StartTime = now
	s.result.StartTimestamp = now.UnixMilli()

	cmd := exec.CommandContext(ctx, s.Command, s.Args...)
	output, err := cmd.CombinedOutput()

	s.result.RoundTripTime.Duration = time.Since(now)

	outputFmt := func(output []byte) string {
		s := string(output)
		if len(strings.TrimSpace(s)) <= 0 {
			return "empty"
		}
		return s
	}
	status := probe.StatusUp
	if err != nil {
		exitCode := 0
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		s.result.Message = fmt.Sprintf("Error: %v, ExitCode(%d), Output:%s", err, exitCode, outputFmt(output))
		log.Errorf(s.result.Message)
		status = probe.StatusDown
	}
	log.Debugf("[%s] - %s", s.Kind(), s.CommandLine())
	log.Debugf("[%s] - %s", s.Kind(), outputFmt(output))

	if err := s.CheckOutput(output); err != nil {
		log.Errorf("[%s] - %v", s.Kind(), err)
		status = probe.StatusDown
	}

	s.result.PreStatus = s.result.Status
	s.result.Status = status

	s.result.DoStat(s.TimeInterval)
	return *s.result
}

// CheckOutput checks the output text,
// - if it contains a configured string then return nil
// - if it does not contain a configured string then return nil
func (s *Shell) CheckOutput(output []byte) error {

	str := string(output)
	if len(s.Contain) > 0 && !strings.Contains(str, s.Contain) {

		return fmt.Errorf("the output does not contain [%s]", s.Contain)
	}

	if len(s.NotContain) > 0 && strings.Contains(str, s.NotContain) {
		return fmt.Errorf("the output contains [%s]", s.NotContain)

	}
	return nil
}

// CommandLine will return the whole command line which includes command and all arguments
func (s *Shell) CommandLine() string {
	result := s.Command
	for _, arg := range s.Args {
		result += " " + arg
	}
	return result
}
