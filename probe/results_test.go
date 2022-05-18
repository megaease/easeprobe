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
	"os"
	"reflect"
	"testing"
)

func TestAll(t *testing.T) {
	r := []Result{*CreateTestResult(), *CreateTestResult(), *CreateTestResult()}
	r[0].Name = "Test1 Name"
	r[1].Name = "Test2 Name"
	r[2].Name = "Test3 Name"

	SetResultsData(r)
	x := GetResultData("Test1 Name")
	if reflect.DeepEqual(x, r[0]) {
		t.Errorf("GetResult(\"Test1 Name\") = %v, expected %v", x, r[0])
	}

	// ensure we dont save or load from '-'
	if err := SaveDataToFile("-"); err != nil {
		t.Errorf("SaveToFile(-) error: %s", err)
	}

	if err := LoadDataFromFile("-"); err != nil {
		t.Errorf("LoadFromFile(-) error: %s", err)
	}

	filename := "/tmp/easeprobe/data.yaml"
	if err := os.MkdirAll("/tmp/easeprobe", 0755); err != nil {
		t.Errorf("Mkdirall(\"/tmp/easeprobe\") error: %v", err)
	}

	if err := SaveDataToFile(filename); err != nil {
		t.Errorf("SaveToFile(%s) error: %s", filename, err)
	}

	if err := LoadDataFromFile(filename); err != nil {
		t.Errorf("LoadFromFile(%s) error: %s", filename, err)
	}

	if reflect.DeepEqual(resultData["Test1 Name"], r[0]) {
		t.Errorf("LoadFromFile(\"%s\") = %v, expected %v", filename, resultData["Test1 Name"], r[0])
	}

	if err := os.RemoveAll("/tmp/easeprobe"); err != nil {
		t.Errorf("RemoveAll(\"/tmp/easeprobe\") = %v, expected nil", err)
	}
}
