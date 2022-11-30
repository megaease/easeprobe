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

package host

import (
	"fmt"
	"strings"
)

// Default threshold
const (
	DefaultCPUThreshold  = 0.8
	DefaultMemThreshold  = 0.8
	DefaultDiskThreshold = 0.95
	DefaultLoadThreshold = 0.8
)

// Threshold is the threshold of a probe
type Threshold struct {
	CPU  float64            `yaml:"cpu,omitempty" json:"cpu,omitempty" jsonschema:"title=CPU threshold,description=CPU threshold (default: 0.8)"`
	Mem  float64            `yaml:"mem,omitempty" json:"mem,omitempty" jsonschema:"title=Memory threshold,description=Memory threshold (default: 0.8)"`
	Disk float64            `yaml:"disk,omitempty" json:"disk,omitempty" jsonschema:"title=Disk threshold,description=Disk threshold (default: 0.95)"`
	Load map[string]float64 `yaml:"load,omitempty" json:"load,omitempty" jsonschema:"title=Load average threshold,description=Load Average M1/M5/M15 threshold (default: 0.8)"`
}

func (t *Threshold) String() string {
	load := []string{}
	for _, v := range t.Load {
		load = append(load, fmt.Sprintf("%.2f", v))
	}

	return fmt.Sprintf("CPU: %.2f, Mem: %.2f, Disk: %.2f, Load: %s", t.CPU, t.Mem, t.Disk, strings.Join(load, "/"))
}
