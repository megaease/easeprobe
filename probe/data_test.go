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

package probe

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var testResults = []Result{
	{
		Name:             "Probe Result - HTTP",
		Endpoint:         "https://www.example.com",
		StartTime:        time.Now().UTC(),
		StartTimestamp:   time.Now().UnixMilli(),
		RoundTripTime:    478783 * time.Microsecond,
		Status:           StatusUp,
		PreStatus:        StatusDown,
		Message:          "Success (http): HTTP Status Code is 200",
		LatestDownTime:   time.Time{},
		RecoveryDuration: 200 * time.Second,
		Stat: Stat{
			Since: time.Now().UTC(),
			Total: 100,
			Status: map[Status]int64{
				StatusUp:   70,
				StatusDown: 30,
			},
			UpTime:        70 * time.Minute,
			DownTime:      30 * time.Minute,
			StatusCounter: *NewStatusCounter(1),
		},
	},
	{
		Name:             "Probe Result - Host",
		Endpoint:         "CPU: 0.15, Mem: 0.10, Disk: 0.90",
		StartTime:        time.Now().UTC(),
		StartTimestamp:   time.Now().UnixMilli(),
		RoundTripTime:    283455 * time.Millisecond,
		Status:           StatusDown,
		PreStatus:        StatusUp,
		Message:          "Error (host/server): CPU Busy! | Memory Shortage! ( CPU: 37.10% - Memory: 49.45% - Disk: 64.00% )",
		LatestDownTime:   time.Time{},
		RecoveryDuration: 5 * time.Minute,
		Stat: Stat{
			Since: time.Now().UTC(),
			Total: 300,
			Status: map[Status]int64{
				StatusInit: 1,
				StatusUp:   270,
				StatusDown: 30,
			},
			UpTime:        270 * time.Minute,
			DownTime:      30 * time.Minute,
			StatusCounter: *NewStatusCounter(2),
		},
	},
}

func newDataFile(file string) error {
	makeAllDir(file)
	SetResultsData(testResults)
	return SaveDataToFile(file)
}

func newDataFileWithOutMeta(file string) error {
	makeAllDir(file)
	SetResultsData(testResults)
	buf, err := yaml.Marshal(resultData)
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, []byte(buf), 0644); err != nil {
		return err
	}
	return nil
}

func isDataFileExisted(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func removeDataFile(file string) {
	os.Remove(file)
}

func makeAllDir(file string) {
	dir, _ := filepath.Split(file)
	os.MkdirAll(dir, 0755)
}

func removeAll(file string) {
	os.RemoveAll(file)
}

func checkData(t *testing.T) {
	for _, r := range testResults {
		assert.Equal(t, r, *resultData[r.Name])
	}
}

func TestNewDataFile(t *testing.T) {
	//disable data file
	newDataFile("-")
	assert.False(t, isDataFileExisted("-"))

	//default data file
	file := global.DefaultDataFile
	newDataFile(file)
	assert.True(t, isDataFileExisted(file))
	removeAll("data/")

	//default data file
	file = "data.yaml"
	newDataFile(file)
	assert.True(t, isDataFileExisted(file))
	removeAll(file)

	//custom data file
	file = "x/y/z/mydata.yaml"
	makeAllDir(file)
	newDataFile(file)
	assert.True(t, isDataFileExisted(file))
	removeAll("x/")

	// errors
	monkey.Patch(os.WriteFile, func(filename string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("error")
	})
	file = "data.yaml"
	err := newDataFile(file)
	assert.Error(t, err)
	removeAll(file)

	monkey.Patch(yaml.Marshal, func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("error")
	})
	err = newDataFile(file)
	assert.Error(t, err)
	removeAll(file)

	monkey.UnpatchAll()
}

func TestLoadDataFile(t *testing.T) {
	// no data file
	file := "data.yaml"
	err := LoadDataFromFile(file)
	assert.True(t, err != nil)

	// create data file
	newDataFile(file)
	if err := LoadDataFromFile(file); err != nil {
		t.Fatal(err)
	}

	assert.True(t, isDataFileExisted(metaData.backup))
	removeAll(metaData.backup)
	checkData(t)

	// errors - backup file
	monkey.Patch(os.Rename, func(oldpath, newpath string) error {
		return fmt.Errorf("error")
	})
	newDataFile(file)
	err = LoadDataFromFile(file)
	assert.Nil(t, err)
	assert.False(t, isDataFileExisted(metaData.backup))
	removeAll(file)

	// errors - unmarshal error
	newDataFile(file)
	monkey.Patch(yaml.Unmarshal, func(in []byte, out interface{}) error {
		return fmt.Errorf("error")
	})
	err = LoadDataFromFile(file)
	assert.Error(t, err)
	removeAll(file)
	monkey.UnpatchAll()

	// errors - decode error
	newDataFile(file)
	var dec *yaml.Decoder
	monkey.PatchInstanceMethod(reflect.TypeOf(dec), "Decode", func(*yaml.Decoder, interface{}) error {
		return fmt.Errorf("error")
	})
	err = LoadDataFromFile(file)
	assert.Error(t, err)
	removeAll(file)

	// errors - read error
	newDataFile(file)
	monkey.Patch(ioutil.ReadFile, func(filename string) ([]byte, error) {
		return nil, fmt.Errorf("error")
	})
	err = LoadDataFromFile(file)
	assert.Error(t, err)
	removeAll(file)

	monkey.UnpatchAll()
}

func numOfBackup(file string) int {
	files, _ := filepath.Glob(file + "-*")
	return len(files)
}

func TestCleanDataFile(t *testing.T) {

	file := "data/data.yaml"
	assert.Equal(t, 0, numOfBackup(file))

	// create data file with backups
	n := 5
	for i := 0; i < n; i++ {
		if err := newDataFile(file); err != nil {
			t.Fatal(err)
		}
		if err := LoadDataFromFile(file); err != nil {
			t.Fatal(err)
		}
		files, _ := filepath.Glob(file + "-*")
		fmt.Printf("i=%d, files=%v\n", i, files)
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, n, numOfBackup(file))

	monkey.Patch(os.Remove, func(filename string) error {
		return fmt.Errorf("error")
	})
	CleanDataFile(file, 3)
	assert.Equal(t, n, numOfBackup(file))

	monkey.UnpatchAll()

	CleanDataFile(file, 10)
	assert.Equal(t, n, numOfBackup(file))

	n = 3
	CleanDataFile(file, n)
	assert.Equal(t, n, numOfBackup(file))

	n = 0
	CleanDataFile(file, n)
	assert.Equal(t, n, numOfBackup(file))

	// clean data file
	removeAll("data/")

	// disable data file
	file = "-"
	CleanDataFile(file, n)
	assert.Equal(t, 0, numOfBackup(file))

	file = "data.yaml"
	CleanDataFile(file, -1)
	assert.Equal(t, 0, numOfBackup(file))

	monkey.Patch(filepath.Glob, func(pattern string) ([]string, error) {
		return nil, fmt.Errorf("error")
	})
	CleanDataFile(file, n)
	assert.Equal(t, 0, numOfBackup(file))

	monkey.UnpatchAll()
}

func TestMetaData(t *testing.T) {
	file := "data/data.yaml"
	// case one: first time to write the data
	newDataFile(file)
	if err := LoadDataFromFile(file); err != nil {
		t.Error(err)
	}
	assert.Equal(t, metaData.Name, global.DefaultProg)
	assert.Equal(t, global.Ver, metaData.Ver)
	removeAll("data/")

	// case two: the data file has no meta
	file = "data/data.yaml"
	newDataFileWithOutMeta(file)
	if err := LoadDataFromFile(file); err != nil {
		t.Error(err)
	}
	assert.Equal(t, metaData.Name, global.DefaultProg)
	assert.Equal(t, metaData.Ver, global.Ver)
	removeAll("data/")

	// case three: the data file has meta
	SetMetaData("myprog", "v1.0.0")
	newDataFile(file)
	if err := LoadDataFromFile(file); err != nil {
		t.Error(err)
	}
	m := GetMetaData()
	assert.Equal(t, m.Name, "myprog")
	assert.Equal(t, m.Ver, global.Ver)
	assert.Equal(t, m.ver, "v1.0.0")

	removeAll("data/")

	// case four: set empty meta
	SetMetaData("", "")
	newDataFile(file)
	err := LoadDataFromFile(file)
	assert.Nil(t, err)
	m = GetMetaData()
	assert.Equal(t, m.Name, global.DefaultProg)
	assert.Equal(t, m.Ver, global.Ver)
	removeAll("data/")

}

type DummyProbe struct {
	MyName     string
	MyResult   *Result
	MyChannels []string
	MyTimeout  time.Duration
	MyInterval time.Duration
}

func (d *DummyProbe) Kind() string {
	return "dummy"
}
func (d *DummyProbe) Name() string {
	return d.MyName
}
func (d *DummyProbe) Channels() []string {
	return d.MyChannels
}
func (d *DummyProbe) Timeout() time.Duration {
	return d.MyTimeout
}
func (d *DummyProbe) Interval() time.Duration {
	return d.MyInterval
}
func (d *DummyProbe) Result() *Result {
	return d.MyResult
}
func (d *DummyProbe) Config(gConf global.ProbeSettings) error {
	return nil
}
func (d *DummyProbe) Probe() Result {
	return *d.MyResult
}

func TestCleanData(t *testing.T) {
	name := "dummy1"
	SetResultData(name, &Result{Name: name})
	name = "dummy2"
	SetResultData(name, &Result{Name: name})
	SetResultsData(testResults)
	assert.Equal(t, len(resultData), 4)

	p := []Prober{
		&DummyProbe{
			MyName:   testResults[0].Name,
			MyResult: &Result{Name: testResults[0].Name},
		},
		&DummyProbe{
			MyName:   testResults[1].Name,
			MyResult: &Result{Name: testResults[1].Name},
		},
		&DummyProbe{
			MyName:   "dummy3",
			MyResult: &Result{Name: "dummy3"},
		},
	}
	CleanData(p)
	assert.Equal(t, len(resultData), 3)
}
