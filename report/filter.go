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

package report

import (
	"fmt"
	"strings"

	"github.com/megaease/easeprobe/probe"

	log "github.com/sirupsen/logrus"
)

// SLAFilter filter the probers
type SLAFilter struct {
	Name       string
	Kind       string
	Endpoint   string
	Status     *probe.Status
	SLAGreater float64
	SLALess    float64
	Message    string
	total      int // the total number of probers
	cnt        int // the number of probers that match the filter
}

// NewEmptyFilter create a new SLAFilter
func NewEmptyFilter() *SLAFilter {
	return &SLAFilter{
		Name:       "",
		Kind:       "",
		Endpoint:   "",
		Status:     nil,
		SLAGreater: 0,
		SLALess:    100,
		Message:    "",
		cnt:        0,
	}
}

// Check check the filter is valid or not
func (f *SLAFilter) Check() error {
	log.Debugf("[Web] Check filter: %+v", f)
	if f.SLAGreater > f.SLALess {
		return fmt.Errorf("Error: Invalid SLA filter: gte(%0.2f) > (%0.2f)", f.SLAGreater, f.SLALess)
	}
	if f.SLAGreater > 100 || f.SLAGreater < 0 {
		return fmt.Errorf("Error: Invalid SLA filter: gte(%0.2f), it must be between 0 - 100", f.SLAGreater)
	}
	if f.SLALess > 100 || f.SLALess < 0 {
		return fmt.Errorf("Error: Invalid SLA filter: lte(%0.2f), it must be between 0 - 100", f.SLALess)
	}
	return nil
}

// HTML return the HTML format string
func (f *SLAFilter) HTML() string {

	span := "<span style=\"font-size:9pt; background-color:#666; color:white; padding:0 5px;border-radius: 3px;\">"
	_span := "</span>  "

	result := ""

	if strings.TrimSpace(f.Name) != "" {
		result += fmt.Sprintf(span+"<b>Name</b>: %s"+_span, f.Name)
	}
	if strings.TrimSpace(f.Kind) != "" {
		result += fmt.Sprintf(span+"<b>Kind</b>: %s"+_span, f.Kind)
	}
	if strings.TrimSpace(f.Endpoint) != "" {
		result += fmt.Sprintf(span+"<b>Endpoint</b>: %s"+_span, f.Endpoint)
	}
	if f.Status != nil {
		result += fmt.Sprintf(span+"<b>Status</b>: %s"+_span, f.Status.String())
	}
	if strings.TrimSpace(f.Message) != "" {
		result += fmt.Sprintf(span+"<b>Message</b>: %s"+_span, f.Message)
	}
	if f.SLAGreater > 0 || f.SLALess < 100 {
		result += fmt.Sprintf(span+"<b>SLA</b>: %.2f%% - %.2f%% "+_span, f.SLAGreater, f.SLALess)
	}

	span = `<span style="font-size:9pt; background-color:#4E944F; color:white; padding:0 5px; margin-left:10px;border-radius: 3px;">`
	result += fmt.Sprintf(span+"<b>%d / %d Probers found!</b>"+_span, f.cnt, f.total)

	result += "<br><br>"
	return result
}

// Filter filter the probers
func (f *SLAFilter) Filter(probers []probe.Prober) []probe.Prober {

	result := make([]probe.Prober, 0)
	for _, p := range probers {
		// if the name is not empty then filter by name
		if strings.TrimSpace(f.Name) != "" && !strings.Contains(p.Name(), f.Name) {
			continue
		}
		// if the kind is not empty then filter by kind
		if strings.TrimSpace(f.Kind) != "" && p.Kind() != f.Kind {
			continue
		}
		// if the endpoint is not empty then filter by endpoint
		if strings.TrimSpace(f.Endpoint) != "" && !strings.Contains(p.Result().Endpoint, f.Endpoint) {
			continue
		}
		// if the status is not right then ignore it
		if f.Status != nil && p.Result().Status != *f.Status {
			continue
		}
		// if the message is not empty then filter by message
		if strings.TrimSpace(f.Message) != "" && !strings.Contains(p.Result().Message, f.Message) {
			continue
		}
		//if the SLA is not right then ignore it
		percent := p.Result().SLAPercent()
		if percent < f.SLAGreater || percent > f.SLALess {
			continue
		}

		result = append(result, p)
	}
	f.total = len(probers)
	f.cnt = len(result)
	return result
}
