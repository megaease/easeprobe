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
	"github.com/megaease/easeprobe/probe"
)

// MergeConstLabels merge const labels from all probers.
// Prometheus requires all metric of the same name have the same set of labels in a collector
func MergeConstLabels(ps []probe.Prober) {
	var constLabels = make(map[string]bool)
	for _, p := range ps {
		for k := range p.LabelMap() {
			constLabels[k] = true
		}
	}

	for _, p := range ps {
		buildConstLabels(p, constLabels)
	}
}

func buildConstLabels(p probe.Prober, constLabels map[string]bool) {
	ls := p.LabelMap()
	if ls == nil {
		ls = make(map[string]string)
		p.SetLabelMap(ls)
	}

	for k := range constLabels {
		if _, ok := ls[k]; !ok {
			ls[k] = ""
		}
	}
}
