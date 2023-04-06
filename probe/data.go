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
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/megaease/easeprobe/global"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// MetaData the meta data of data file
type MetaData struct {
	Name   string `yaml:"name"`
	Ver    string `yaml:"version"`
	ver    string `yaml:"-"` // the version in data file
	file   string `yaml:"-"` // the current file name
	backup string `yaml:"-"` // the backup file name
}

var (
	resultData = map[string]*Result{}
	metaData   = MetaData{
		Name: global.DefaultProg,
		Ver:  global.Ver,
	}
	metaBuf []byte
	mutex   = &sync.RWMutex{}
)

const split = "---\n"

// GetMetaData get the meta data
func GetMetaData() *MetaData {
	return &metaData
}

// SetResultData set the result of probe
// Note: this function would be called by status update goroutine
//
//	int saveData() in cmd/easeprobe/report.go
func SetResultData(name string, result *Result) {
	r := result.Clone()
	mutex.Lock()
	resultData[name] = &r
	mutex.Unlock()

}

// SetResultsData set the results of probe
func SetResultsData(r []Result) {
	for i := 0; i < len(r); i++ {
		SetResultData(r[i].Name, &r[i])
	}
}

// GetResultData get the result of probe
// Note: the function would be called by Data Saving, SLA Report, Web Server
func GetResultData(name string) *Result {
	mutex.RLock()
	defer mutex.RUnlock()
	if v, ok := resultData[name]; ok {
		r := v.Clone()
		return &r
	}
	return nil
}

// CleanData removes the items in resultData not in []Prober
// Note: No need to consider the thread-safe, because this function is only called once during the startup
func CleanData(p []Prober) {
	var data = map[string]*Result{}
	for i := 0; i < len(p); i++ {
		r := p[i].Result()
		d := GetResultData(r.Name)
		if d != nil {
			data[r.Name] = d
		} else {
			data[r.Name] = r
		}
	}
	resultData = data
}

// SaveDataToFile save the results to file
// Note: No need to consider the thread-safe, because this function and SetResultData in same goroutine
func SaveDataToFile(filename string) error {
	metaData.file = filename
	if strings.TrimSpace(filename) == "-" {
		return nil
	}

	dataBuf, err := yaml.Marshal(resultData)
	if err != nil {
		return err
	}

	genMetaBuf()
	buf := append(metaBuf, dataBuf...)

	if err := os.WriteFile(filename, []byte(buf), 0644); err != nil {
		return err
	}
	return nil
}

// LoadDataFromFile load the results from file
// Note: No need to consider the thread-safe, because this function is only called once during the startup
func LoadDataFromFile(filename string) error {

	// if the data file is disabled, return
	if strings.TrimSpace(filename) == "-" {
		return nil
	}

	// if the data file is not exist, return
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Multiple YAML Documents reading
	dec := yaml.NewDecoder(bytes.NewReader(buf))
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if v, ok := value.(map[string]interface{}); ok {
			valueBytes, _ := yaml.Marshal(value)
			if v["version"] != nil {
				if err := yaml.Unmarshal(valueBytes, &metaData); err != nil {
					log.Warnf("Load meta data error: %v", err)
				} else {
					log.Debugf("Load meta data: name[%s], version[%s]", metaData.Name, metaData.Ver)
				}
			} else {
				if err := yaml.Unmarshal(valueBytes, &resultData); err != nil {
					return err
				}
			}
		}
	}

	// set the meta name and version
	// - if the Name is found in the data file, use it, otherwise use the default
	// - always use the program version for the data file.
	metaData.ver = metaData.Ver // save the file's version
	SetMetaData(metaData.Name, global.Ver)

	// backup the current data file
	time := time.Now().UTC().Format(time.RFC3339Nano)
	// replace ":" to "_" for windows platform compliance
	time = strings.Replace(time, ":", "_", -1)
	metaData.backup = filename + "-" + time
	if err := os.Rename(filename, metaData.backup); err != nil {
		log.Warnf("Backup data file error: %v", err)
	}

	return nil
}

// CleanDataFile keeps the max backup of data file
// Note: No need to consider the thread-safe, because this function is only called once during the startup
func CleanDataFile(filename string, backups int) {
	if strings.TrimSpace(filename) == "-" {
		return
	}

	// if backups is negative value, keep all backup files
	if backups < 0 {
		return
	}

	// get all of the backup files
	pattern := filename + "-*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Errorf("Cannot clean data file: %v", err)
		return
	}

	// if backups is not exceed the max number of backup files, return
	if len(matches) <= backups {
		log.Debugf("No need to clean data file (%d - %d) ", backups, len(matches))
		return
	}

	// remove the oldest backup files
	sort.Strings(matches)

	for i := 0; i < len(matches)-backups; i++ {
		if err := os.Remove(matches[i]); err != nil {
			log.Errorf("Cannot clean data file: %v", err)
			continue
		}
		log.Infof("Clean data file: %s", matches[i])
	}
}

// SetMetaData set the meta data
// Note: No need to consider the thread-safe, because this function is only called during the startup
func SetMetaData(name string, ver string) {

	metaData.Name = name
	metaData.Ver = ver

	// reconstructure the meta buf
	genMetaBuf()
}

func genMetaBuf() {
	// if the meta data is not exist in current data file, using the default.
	if metaData.Name == "" {
		metaData.Name = global.DefaultProg
	}
	if metaData.Ver == "" {
		metaData.Ver = global.Ver
	}
	metaBuf, _ = yaml.Marshal(metaData)
	metaBuf = append([]byte(split), metaBuf...)
	metaBuf = append(metaBuf, []byte(split)...)
}
