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
	"regexp"
	"strings"
)

// TextChecker is the struct to check the output
type TextChecker struct {
	Contain    string `yaml:"contain,omitempty" json:"contain,omitempty" jsonschema:"title=Contain Text,description=the string must be contained"`
	NotContain string `yaml:"not_contain,omitempty" json:"not_contain,omitempty" jsonschema:"title=Not Contain Text,description=the string must not be contained"`
	RegExp     bool   `yaml:"regex,omitempty" json:"regex,omitempty" jsonschema:"title=regex,description=use regular expression to check the contain or not contain"`

	containReg    *regexp.Regexp `yaml:"-" json:"-"`
	notContainReg *regexp.Regexp `yaml:"-" json:"-"`
}

// Config the text checker initialize the regexp
func (tc *TextChecker) Config() (err error) {
	if !tc.RegExp {
		return nil
	}

	if len(tc.Contain) == 0 {
		tc.containReg = nil
	} else if tc.containReg, err = regexp.Compile(tc.Contain); err != nil {
		tc.containReg = nil
		return err
	}

	if len(tc.NotContain) == 0 {
		tc.notContainReg = nil
	} else if tc.notContainReg, err = regexp.Compile(tc.NotContain); err != nil {
		tc.notContainReg = nil
		return err
	}

	return nil
}

// Check the text
func (tc *TextChecker) Check(Text string) error {
	if tc.RegExp {
		return tc.CheckRegExp(Text)
	}
	return tc.CheckText(Text)
}

func (tc *TextChecker) String() string {
	if tc.RegExp {
		return fmt.Sprintf("RegExp Mode - Contain:[%s], NotContain:[%s]", tc.Contain, tc.NotContain)
	}
	return fmt.Sprintf("Text Mode - Contain:[%s], NotContain:[%s]", tc.Contain, tc.NotContain)
}

// CheckText checks the output text,
// - if it contains a configured string then return nil
// - if it does not contain a configured string then return nil
func (tc *TextChecker) CheckText(Output string) error {

	if len(tc.Contain) > 0 && !strings.Contains(Output, tc.Contain) {
		return fmt.Errorf("the output does not contain [%s]", tc.Contain)
	}

	if len(tc.NotContain) > 0 && strings.Contains(Output, tc.NotContain) {
		return fmt.Errorf("the output contains [%s]", tc.NotContain)
	}
	return nil
}

// CheckRegExp checks the output text,
// - if it contains a configured pattern then return nil
// - if it does not contain a configured pattern then return nil
func (tc *TextChecker) CheckRegExp(Output string) error {

	if len(tc.Contain) > 0 && tc.containReg != nil && !tc.containReg.MatchString(Output) {
		return fmt.Errorf("the output does not match the pattern [%s]", tc.Contain)
	}

	if len(tc.NotContain) > 0 && tc.notContainReg != nil && tc.notContainReg.MatchString(Output) {
		return fmt.Errorf("the output match the pattern [%s]", tc.NotContain)
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
