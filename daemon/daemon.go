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

// Package daemon is the daemon implementation.
package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/megaease/easeprobe/global"
)

// Config is the daemon config
type Config struct {
	PIDFile string
	pidFd   *os.File
}

// NewPIDFile create a new pid file
func NewPIDFile(pidfile string) (*Config, error) {
	if pidfile == "" {
		return nil, fmt.Errorf("pid file is empty")
	}

	fi, err := os.Stat(pidfile)

	if err == nil { // file exists
		if fi.IsDir() {
			pidfile = filepath.Join(pidfile, global.DefaultPIDFile)
		}
		li, _ := os.Lstat(pidfile)
		if li != nil && (li.Mode()&os.ModeSymlink == os.ModeSymlink) {
			os.Remove(pidfile)
		}
	} else if errors.Is(err, os.ErrNotExist) { // file not exists
		// create all of directories
		if e := os.MkdirAll(filepath.Dir(pidfile), 0755); e != nil {
			return nil, e
		}
	} else {
		return nil, err
	}

	c := &Config{
		PIDFile: pidfile,
	}

	pidstr := fmt.Sprintf("%d", os.Getpid())
	if err := os.WriteFile(c.PIDFile, []byte(pidstr), 0600); err != nil {
		return nil, err
	}

	c.pidFd, _ = os.OpenFile(c.PIDFile, os.O_APPEND|os.O_EXCL, 0600)
	return c, nil
}

// CheckPIDFile check if the pid file exists
// if the PID file exists, return the PID of the process
// if the PID file does not exist, return -1
func (c *Config) CheckPIDFile() (int, error) {
	buf, err := os.ReadFile(c.PIDFile)
	if err != nil {
		return -1, nil
	}

	pidstr := strings.TrimSpace(string(buf))
	pid, err := strconv.Atoi(pidstr)
	if err != nil {
		return -1, nil
	}

	if processExists(pid) {
		return pid, fmt.Errorf("pid file(%s) found, ensure %s(%d) is not running",
			c.PIDFile, global.DefaultProg, pid)
	}

	return -1, nil
}

// RemovePIDFile remove the pid file
func (c *Config) RemovePIDFile() error {
	c.pidFd.Close()
	return os.Remove(c.PIDFile)
}
