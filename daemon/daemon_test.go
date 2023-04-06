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

package daemon

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"

	"github.com/stretchr/testify/assert"
)

func testPIDFile(pidfile string, t *testing.T) {
	c, err := NewPIDFile(pidfile)
	if err != nil {
		t.Fatalf("Could not create the pid file, %v", err)
	}

	_, err = c.CheckPIDFile()
	if err == nil {
		t.Fatalf("Could not found the pid file, %v", err)
	}

	if err := c.RemovePIDFile(); err != nil {
		t.Fatalf("Could not remove the pid file, %v", err)
	}
}

func TestPIDFileNotExist(t *testing.T) {
	pidfile := filepath.Join(global.GetWorkDir(), global.DefaultPIDFile)
	testPIDFile(pidfile, t)
}

func TestPIDFileExist(t *testing.T) {
	path := filepath.Join(global.GetWorkDir(), global.DefaultPIDFile)
	os.WriteFile(path, []byte("1"), 0644)
	testPIDFile(path, t)
}

func TestPIDFileDir(t *testing.T) {
	pidfile := global.GetWorkDir()
	testPIDFile(pidfile, t)
}

func TestPIDFileSymLink(t *testing.T) {
	path := filepath.Join(global.GetWorkDir(), "test")
	target := "test.txt"
	os.MkdirAll(path, 0755)
	os.WriteFile(filepath.Join(path, "test.txt"), []byte("Hello\n"), 0644)
	symlink := filepath.Join(path, "easeprobe.pid")
	os.Symlink(target, symlink)

	c, err := NewPIDFile(symlink)
	if err != nil {
		t.Fatalf("Could not create the pid file, %v", err)
	}

	_, err = c.CheckPIDFile()
	if err == nil {
		t.Fatalf("Could not found the pid file, %v", err)
	}

	buf, err := ioutil.ReadFile(filepath.Join(path, "test.txt"))
	if err != nil {
		t.Fatalf("Could not read the pid file, %v", err)
	}

	assert.Equal(t, "Hello\n", string(buf))

	if err := c.RemovePIDFile(); err != nil {
		t.Fatalf("Could not remove the pid file, %v", err)
	}

	os.RemoveAll(path)
}

func TestPIDFileFailed(t *testing.T) {
	file := ""
	conf, err := NewPIDFile(file)
	assert.Nil(t, conf)
	assert.NotNil(t, err)

	file = "./"
	conf, err = NewPIDFile(file)
	assert.FileExists(t, global.DefaultPIDFile)
	os.RemoveAll(global.DefaultPIDFile)

	file = "dir/easedprobe.pid"
	conf, err = NewPIDFile(file)
	assert.FileExists(t, file)
	conf.RemovePIDFile()
	os.RemoveAll("dir")

	monkey.Patch(os.WriteFile, func(string, []byte, os.FileMode) error {
		return fmt.Errorf("error")
	})

	conf, err = NewPIDFile(file)
	assert.Nil(t, conf)
	assert.NotNil(t, err)
	assert.DirExists(t, "dir")
	assert.NoFileExists(t, file)
	os.RemoveAll("dir")

	monkey.Patch(os.MkdirAll, func(string, os.FileMode) error {
		return fmt.Errorf("error")
	})
	conf, err = NewPIDFile(file)
	assert.Nil(t, conf)
	assert.NotNil(t, err)
	assert.NoFileExists(t, file)
	assert.NoDirExists(t, "dir")

	monkey.Patch(os.Stat, func(string) (os.FileInfo, error) {
		return nil, fmt.Errorf("error")
	})
	conf, err = NewPIDFile(file)
	assert.Nil(t, conf)
	assert.NotNil(t, err)
	assert.NoFileExists(t, file)
	assert.NoDirExists(t, "dir")

	monkey.UnpatchAll()
}

func TestCheckPIDFile(t *testing.T) {
	file := "dir/easedprobe.pid"
	conf, err := NewPIDFile(file)
	assert.Nil(t, err)
	assert.FileExists(t, file)

	pid, err := conf.CheckPIDFile()
	assert.Equal(t, os.Getpid(), pid)
	assert.NotNil(t, err)

	monkey.Patch(processExists, func(int) bool {
		return false
	})
	pid, err = conf.CheckPIDFile()
	assert.Equal(t, -1, pid)
	assert.Nil(t, err)

	os.WriteFile(conf.PIDFile, []byte("invalid pid"), 0644)
	pid, err = conf.CheckPIDFile()
	assert.Equal(t, -1, pid)
	assert.Nil(t, err)

	conf.RemovePIDFile()
	os.RemoveAll("dir")

	pid, err = conf.CheckPIDFile()
	assert.Equal(t, -1, pid)
	assert.Nil(t, err)

	monkey.UnpatchAll()
}
