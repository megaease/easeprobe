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

package monkey

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type myStruct struct{}

func (s *myStruct) Method() string {
	return "original"
}

func TestPatch(t *testing.T) {
	assert.Equal(t, "original", strings.Clone("original"))

	Patch(strings.Clone, func(s string) string { return "replacement" })
	assert.Equal(t, "replacement", strings.Clone("original"))

	Unpatch(strings.Clone)
	assert.Equal(t, "original", strings.Clone("original"))
}

func TestPatchInstanceMethod(t *testing.T) {
	assert.Equal(t, "original", (&myStruct{}).Method())

	PatchInstanceMethod(reflect.TypeOf(&myStruct{}), "Method", func(*myStruct) string { return "replacement" })
	assert.Equal(t, "replacement", (&myStruct{}).Method())

	UnpatchInstanceMethod(reflect.TypeOf(&myStruct{}), "Method")
	assert.Equal(t, "original", (&myStruct{}).Method())
}

func TestUnpatchAll(t *testing.T) {
	assert.Equal(t, "original", strings.Clone("original"))
	Patch(strings.Clone, func() string { return "replacement" })
	assert.Equal(t, "replacement", strings.Clone("original"))

	assert.Equal(t, "original", (&myStruct{}).Method())
	PatchInstanceMethod(reflect.TypeOf(&myStruct{}), "Method", func(*myStruct) string { return "replacement" })
	assert.Equal(t, "replacement", (&myStruct{}).Method())

	UnpatchAll()
	assert.Equal(t, "original", strings.Clone("original"))
	assert.Equal(t, "original", (&myStruct{}).Method())
}
