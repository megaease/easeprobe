package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	log "github.com/sirupsen/logrus"
)

// Kind is the type of probe
const Kind string = "ssh"

// SSH implements a config for ssh command (os.Exec)
type SSH struct {
	base.DefaultOptions `yaml:",inline"`
	PrivateKey          string   `yaml:"key"`
	Host                string   `yaml:"host"`
	User                string   `yaml:"username"`
	Password            string   `yaml:"password"`
	Command             string   `yaml:"cmd"`
	Args                []string `yaml:"args,omitempty"`
	Env                 []string `yaml:"env,omitempty"`
	Contain             string   `yaml:"contain,omitempty"`
	NotContain          string   `yaml:"not_contain,omitempty"`
}

// Config SSH Config Object
func (s *SSH) Config(gConf global.ProbeSettings) error {
	kind := "ssh"
	tag := ""
	name := s.ProbeName
	s.DefaultOptions.Config(gConf, kind, tag, name, probe.CommandLine(s.Command, s.Args), s.DoProbe)

	if len(s.Password) <= 0 && len(s.PrivateKey) <= 0 {
		return fmt.Errorf("password or private key is required")
	}

	log.Debugf("[%s] configuration: %+v, %+v", s.ProbeKind, s, s.Result())
	return nil
}

// DoProbe return the checking result
func (s *SSH) DoProbe() (bool, string) {

	output, err := s.RunSSHCmd()

	status := true
	message := "SSH Command has been Run Successfully!"

	if err != nil {
		log.Errorf("[%s / %s] %v", s.ProbeKind, s.ProbeName, err)
		status = false
		message = err.Error() + " - " + output
	}

	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CommandLine(s.Command, s.Args))
	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CheckEmpty(string(output)))

	if err := probe.CheckOutput(s.Contain, s.NotContain, string(output)); err != nil {
		log.Errorf("[%s / %s] - %v", s.ProbeKind, err)
		message = fmt.Sprintf("Error: %v", err)
		status = false
	}

	return status, message
}

// RunSSHCmd run ssh command
func (s *SSH) RunSSHCmd() (string, error) {

	var Auth []ssh.AuthMethod

	if len(s.Password) > 0 {
		Auth = append(Auth, ssh.Password(s.Password))
	}

	var hostKeyCallback ssh.HostKeyCallback

	if len(s.PrivateKey) > 0 {
		key, err := ioutil.ReadFile(s.PrivateKey)
		if err != nil {
			return "", err
		}

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return "", err
		}

		Auth = append(Auth, ssh.PublicKeys(signer))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		hostKeyCallback = nil
		log.Warnf("[%s / %s] unable to get home directory: %v", s.ProbeKind, s.ProbeName, err)
	} else {
		hostKeyCallback, err = knownhosts.New(home + "/.ssh/known_hosts")
		if err != nil {
			log.Warnf("[%s / %s] could not create hostkeycallback function: ", s.ProbeKind, s.ProbeName, err)
		}
	}

	config := &ssh.ClientConfig{
		User:            s.User,
		Auth:            Auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         s.Timeout(),
	}

	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", s.Host, config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Create a session.
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Set up environment variables
	env := ""
	for _, e := range s.Env {
		env += "export " + e + ";"
	}

	// Creating the buffer which will hold the remotely executed command's output.
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf
	if err := session.Run(env + probe.CommandLine(s.Command, s.Args)); err != nil {
		return stderrBuf.String(), err
	}

	return stdoutBuf.String(), nil

}
