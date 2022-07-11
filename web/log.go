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

// The following Source Code comes from
// https://github.com/go-chi/chi/blob/master/_examples/logging/main.go

package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/megaease/easeprobe/global"
	"github.com/sirupsen/logrus"
)

// AccessLog is the default access log format
type AccessLog struct {
	RemoteAddr string `json:"remote_addr"`
	UserID     string `json:"user_id"`
	Time       string `json:"time"`
	Method     string `json:"method"`
	Request    string `json:"request"`
	Status     string `json:"status"`
	Bytes      string `json:"bytes"`
	Elapsed    string `json:"elapsed"`
	Referrer   string `json:"referrer"`
	UserAgent  string `json:"user_agent"`
	Stack      string `json:"stack"`
	Panic      string `json:"panic"`
}

func (a AccessLog) String() string {
	if a.Panic != "" {
		return fmt.Sprintf("%s %s \"%s\" %s %s %s %s %s %s \"%s\" %s %s",
			a.RemoteAddr, a.UserID, a.Time, a.Method, a.Request, a.Status, a.Bytes, a.Elapsed, a.Referrer, a.UserAgent, a.Stack, a.Panic)
	}
	return fmt.Sprintf("%s %s \"%s\" %s %s %s %s %s %s \"%s\"",
		a.RemoteAddr, a.UserID, a.Time, a.Method, a.Request, a.Status, a.Bytes, a.Elapsed, a.Referrer, a.UserAgent)
}

// PlainFormatter is a plain text formatter
type PlainFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

// Format formats the log entry
func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
	return []byte(fmt.Sprintf("%s %s\n", timestamp, entry.Message)), nil
}

// StructuredLogger is a simple, but powerful implementation of a custom structured
// logger backed on logrus.
type StructuredLogger struct {
	Logger *logrus.Logger
}

// NewStructuredLogger returns a new StructuredLogger
func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	plainFormatter := new(PlainFormatter)
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05"
	plainFormatter.LevelDesc = []string{"PANC", "FATL", "ERRO", "WARN", "INFO", "DEBG"}
	logger.SetFormatter(plainFormatter)
	return middleware.RequestLogger(&StructuredLogger{logger})
}

// NewLogEntry creates a new logrus.Entry from the request
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {

	access := AccessLog{}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		access.UserID = reqID
	}
	access.Method = r.Method
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	access.Request = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	access.RemoteAddr = r.RemoteAddr
	access.UserAgent = r.UserAgent()
	access.Time = time.Now().In(global.GetTimeLocation()).Format(time.RFC1123)

	access.Referrer = r.Referer()

	entry := &StructuredLoggerEntry{l.Logger, access}
	return entry
}

// StructuredLoggerEntry is a logrus.Entry with some additional fields
type StructuredLoggerEntry struct {
	Logger    *logrus.Logger
	AccessLog AccessLog
}

// Write is method for Status Bytes and Elapsed
func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.AccessLog.Status = fmt.Sprintf("%d", status)
	l.AccessLog.Bytes = fmt.Sprintf("%d", bytes)
	l.AccessLog.Elapsed = fmt.Sprintf("%0.3fms", float64(elapsed.Nanoseconds())/1000000.0)

	l.Logger.Infoln(l.AccessLog)
}

// Panic is a convenience method for Logrus
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {

	l.AccessLog.Panic = fmt.Sprintf("%+v", v)
	l.AccessLog.Stack = string(stack)
}
