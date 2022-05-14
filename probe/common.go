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
	"strings"
)

// CommandLine will return the whole command line which includes command and all arguments
func CommandLine(cmd string, args []string) string {
	result := cmd
	for _, arg := range args {
		result += " " + arg
	}
	return result
}

// CheckOutput checks the output text,
// - if it contains a configured string then return nil
// - if it does not contain a configured string then return nil
func CheckOutput(Contain, NotContain, Output string) error {

	if len(Contain) > 0 && !strings.Contains(Output, Contain) {
		return fmt.Errorf("the output does not contain [%s]", Contain)
	}

	if len(NotContain) > 0 && strings.Contains(Output, NotContain) {
		return fmt.Errorf("the output contains [%s]", NotContain)
	}
	return nil
}

// CheckEmpty return "empty" if the string is empty
func CheckEmpty(s string) string {
	if len(strings.TrimSpace(s)) <= 0 {
		return "empty"
	}
	return s
}
