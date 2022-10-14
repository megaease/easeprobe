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

// StatusHistory is a history of status
type StatusHistory struct {
	Status  bool
	Message string
}

// StatusCounter is the object to count the status
type StatusCounter struct {
	StatusHistory []StatusHistory // the status history
	MaxLen        int             // the max length of the status history
	CurrentStatus bool            // the current status
	StatusCount   int             // the count of the same status
}

// NewStatusCounter return a StatusCounter object
func NewStatusCounter(maxLen int) *StatusCounter {
	threshold := &StatusCounter{
		StatusHistory: make([]StatusHistory, 0),
		MaxLen:        maxLen,
		CurrentStatus: true,
		StatusCount:   0,
	}
	return threshold
}

// AppendStatus appends the status
func (s *StatusCounter) AppendStatus(status bool, message string) {

	if status != s.CurrentStatus { // status change, reset the status count
		s.StatusCount = 0
		s.CurrentStatus = status
	}
	if s.StatusCount < s.MaxLen { // count the status if it is less than the max length
		s.StatusCount++
	}

	h := StatusHistory{
		Status:  status,
		Message: message,
	}
	// append the status
	s.StatusHistory = append(s.StatusHistory, h)

	// pop up the first element
	if len(s.StatusHistory) > s.MaxLen {
		s.StatusHistory = s.StatusHistory[1:]
	}
}

// SetMaxLen sets the max length of the status history
func (s *StatusCounter) SetMaxLen(maxLen int) {
	s.MaxLen = maxLen
	if len(s.StatusHistory) > s.MaxLen {
		s.StatusHistory = s.StatusHistory[len(s.StatusHistory)-s.MaxLen:]
	}
}

// Clone returns a copy of the StatusThreshold
func (s *StatusCounter) Clone() StatusCounter {
	return StatusCounter{
		StatusHistory: s.StatusHistory,
		MaxLen:        s.MaxLen,
		CurrentStatus: s.CurrentStatus,
		StatusCount:   s.StatusCount,
	}
}
