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
}
