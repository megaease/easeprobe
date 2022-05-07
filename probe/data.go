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
	"time"

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

	dir, _ := filepath.Split(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(buf)
	return nil
}

// LoadDataFromFile load the results from file
func LoadDataFromFile(filename string) error {
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

	time := time.Now().Format(time.RFC3339)
	os.Rename(filename, filename+"-"+time)

	return nil
}
