package ssh

import (
	"io/ioutil"
	"net"

	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Endpoint is SSH Endpoint
type Endpoint struct {
	PrivateKey string      `yaml:"key"`
	Host       string      `yaml:"host"`
	User       string      `yaml:"username"`
	Password   string      `yaml:"password"`
	client     *ssh.Client `yaml:"-"`
}

// ParseHost check the host is configured the port or not
func (e *Endpoint) ParseHost() error {

	if strings.LastIndex(e.Host, ":") < 0 {
		e.Host = e.Host + ":22"
	}
	if strings.Index(e.Host, "@") > 0 {
		e.User = e.Host[:strings.Index(e.Host, "@")]
		e.Host = e.Host[strings.Index(e.Host, "@")+1:]
	}
	_, _, err := net.SplitHostPort(e.Host)

	if err != nil {
		return err
	}
	return nil
}

// SSHConfig returns the ssh.ClientConfig
func (e *Endpoint) SSHConfig(kind, name string, timeout time.Duration) (*ssh.ClientConfig, error) {
	var Auth []ssh.AuthMethod

	if len(e.Password) > 0 {
		Auth = append(Auth, ssh.Password(e.Password))
	}

	if len(e.PrivateKey) > 0 {
		key, err := ioutil.ReadFile(e.PrivateKey)
		if err != nil {
			return nil, err
		}

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		Auth = append(Auth, ssh.PublicKeys(signer))
	}

	config := &ssh.ClientConfig{
		User:            e.User,
		Auth:            Auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	return config, nil
}
