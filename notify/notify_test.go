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

package notify

import (
	"reflect"
	"testing"

	"github.com/megaease/easeprobe/notify/base"
	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	n := &base.DefaultNotify{}
	assert.Implements(t, (*Notify)(nil), n)

	v := reflect.ValueOf(Config{})
	for i := 0; i < v.NumField(); i++ {
		assert.IsType(t, reflect.Slice, v.Field(i).Kind())
		n := reflect.TypeOf(v.Field(i).Interface()).Elem()
		assert.Implements(t, (*Notify)(nil), reflect.New(n).Interface())
	}
}
