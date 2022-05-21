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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var resultData = map[string]*Result{}

// SetResultData set the result of probe
func SetResultData(name string, result *Result) {
	resultData[name] = result
}

// SetResultsData set the results of probe
func SetResultsData(r []Result) {
	for i := 0; i < len(r); i++ {
		SetResultData(r[i].Name, &r[i])
	}
}

// GetResultData get the result of probe
func GetResultData(name string) *Result {
	if v, ok := resultData[name]; ok {
		return v
	}
	return nil
}

// CleanData removes the items in resultData not in r[]
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
func SaveDataToFile(filename string) error {
	buf, err := yaml.Marshal(resultData)
	if err != nil {
		return err
	}

	if strings.TrimSpace(filename) == "-" {
		return nil
	}

	if err := ioutil.WriteFile(filename, buf, 0644); err != nil {
		return err
	}
	return nil
}

// LoadDataFromFile load the results from file
func LoadDataFromFile(filename string) error {
	if strings.TrimSpace(filename) == "-" {
		return nil
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(buf, &resultData)
	if err != nil {
		return err
	}

	time := time.Now().UTC().Format(time.RFC3339)
	os.Rename(filename, filename+"-"+time)

	return nil
}

// CleanDataFile keeps the max backup of data file
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
