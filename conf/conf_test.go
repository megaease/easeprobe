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

package conf

import (
	"io/ioutil"
	"os"
	"testing"
)

func testisExternalURL(url string, expects bool, t *testing.T) {
	if got := isExternalURL(url); got != expects {
		t.Errorf("isExternalURL(\"%s\") = %v, expected %v", url, got, expects)
	}
}

func TestPathAndURL(t *testing.T) {
	testisExternalURL("/tmp", false, t)
	testisExternalURL("//tmp", false, t)
	testisExternalURL("file:///tmp", false, t)
	testisExternalURL("http://", false, t)
	testisExternalURL("https://", false, t)
	testisExternalURL("hTtP://", false, t)
	testisExternalURL("http", false, t)
	testisExternalURL("https", false, t)
	testisExternalURL("ftp", false, t)
	testisExternalURL("hTtP://127.0.0.1", true, t)
	testisExternalURL("localhost", false, t)
	testisExternalURL("ftp://127.0.0.1", false, t)
}

func TestGetYamlFileFromFile(t *testing.T) {
	//content := []byte("temporary file's content")

	if _, err := getYamlFileFromFile("/tmp/nonexistent"); err == nil {
		t.Errorf("getYamlFileFromFile(\"/tmp/nonexistent\") = nil, expected error")
	}

	tmpfile, err := ioutil.TempFile("", "invalid*.yaml")
	if err != nil {
		t.Errorf("TempFile(\"invalid*.yaml\") %v", err)
	}

	defer os.Remove(tmpfile.Name()) // clean up

	// test empty file
	data, err := getYamlFileFromFile(tmpfile.Name())
	if err != nil {
		t.Errorf("getYamlFileFromFile(\"%s\") = %v, expected nil", tmpfile.Name(), err)
	}
	if string(data) != "" {
		t.Errorf("getYamlFileFromFile(\"%s\") got data %s, expected nil", tmpfile.Name(), data)
	}
}
