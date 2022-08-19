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

// Package log is the log package for easeprobe.
package log

import (
	"fmt"
	"time"

	"github.com/megaease/easeprobe/global"

	log "github.com/sirupsen/logrus"
)

// SysLogFormatter is log custom format
type SysLogFormatter struct {
	Type Type `yaml:"-"`
}

// Format details
func (s *SysLogFormatter) Format(entry *log.Entry) ([]byte, error) {
	if s.Type == SysLog {
		return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
	}

	timestamp := time.Now().Local().Format(time.RFC3339)
	host := global.GetEaseProbe().Host
	app := global.GetEaseProbe().Name

	msg := fmt.Sprintf("%s %s %s %s %s\n", timestamp, host, app, entry.Level.String(), entry.Message)
	return []byte(msg), nil
}
