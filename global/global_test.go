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

package global

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWritableDir(t *testing.T) {
	filename := ""
	dir := MakeDirectory(filename)
	assert.Equal(t, GetWorkDir(), dir)

	filename = "./test.txt"
	dir = MakeDirectory(filename)
	exp, _ := filepath.Abs(filename)
	assert.Equal(t, exp, dir)

	filename = "./none/existed/test.txt"
	exp, _ = filepath.Abs(filename)
	dir = MakeDirectory(filename)
	os.RemoveAll("./none")
	assert.Equal(t, exp, dir)

	filename = "~/none/existed/test.txt"
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	exp = filepath.Join(home, "none/existed/test.txt")
	dir = MakeDirectory(filename)
	os.RemoveAll(home + "/none")
	assert.Equal(t, exp, dir)
}
