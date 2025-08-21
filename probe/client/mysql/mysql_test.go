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

package mysql

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/monkey"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/stretchr/testify/assert"
)

func TestMySQL(t *testing.T) {
	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.MySQL,
		Username:   "username",
		Password:   "password",
		TLS: global.TLS{
			CA:   "ca",
			Cert: "cert",
			Key:  "key",
		},
	}

	my, err := New(conf)
	assert.Nil(t, my)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "TLS Config Error")

	conf.TLS = global.TLS{}
	my, err = New(conf)
	assert.Nil(t, err)
	assert.Equal(t, "MySQL", my.Kind())
	connStr := fmt.Sprintf("%s:%s@tcp(%s)/?timeout=%s",
		conf.Username, conf.Password, conf.Host, conf.Timeout().Round(time.Second))
	assert.Equal(t, connStr, my.ConnStr)

	conf.Password = ""
	my, err = New(conf)
	connStr = fmt.Sprintf("%s@tcp(%s)/?timeout=%s",
		conf.Username, conf.Host, conf.Timeout().Round(time.Second))
	assert.Equal(t, connStr, my.ConnStr)

	monkey.Patch(sql.Open, func(driverName, dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	})
	var db *sql.DB
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Close", func(_ *sql.DB) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Ping", func(_ *sql.DB) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Query", func(_ *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
		return &sql.Rows{}, nil
	})
	var r *sql.Rows
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Close", func(_ *sql.Rows) error {
		return nil
	})

	s, m := my.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	// TLS config success
	var tc *global.TLS
	monkey.PatchInstanceMethod(reflect.TypeOf(tc), "Config", func(_ *global.TLS) (*tls.Config, error) {
		return &tls.Config{}, nil
	})
	monkey.Patch(mysql.RegisterTLSConfig, func(name string, config *tls.Config) error {
		return nil
	})

	my, err = New(conf)
	assert.NotNil(t, my.tls)
	s, m = my.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	//  Query error
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Query", func(_ *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
		return nil, fmt.Errorf("query error")
	})
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "query error")

	// Ping error
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Ping", func(_ *sql.DB) error {
		return fmt.Errorf("ping error")
	})
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "ping error")

	// Connect error
	monkey.Patch(sql.Open, func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, fmt.Errorf("connect error")
	})
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "connect error")

	monkey.UnpatchAll()
}

func TestMySQLPasswordEncoding(t *testing.T) {
	// Test case for issue #673: passwords with special characters should be URL encoded
	testCases := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{
			name:     "Simple password",
			username: "root",
			password: "password123",
			expected: "root:password123@tcp(localhost:3306)/?timeout=0s",
		},
		{
			name:     "Password with dollar sign",
			username: "root",
			password: "AB10$CCC123",
			expected: "root:AB10%24CCC123@tcp(localhost:3306)/?timeout=0s",
		},
		{
			name:     "Password with special characters",
			username: "user@domain",
			password: "pass@word#123!",
			expected: "user%40domain:pass%40word%23123%21@tcp(localhost:3306)/?timeout=0s",
		},
		{
			name:     "Empty password",
			username: "root",
			password: "",
			expected: "root@tcp(localhost:3306)/?timeout=0s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := conf.Options{
				Host:       "localhost:3306",
				DriverType: conf.MySQL,
				Username:   tc.username,
				Password:   tc.password,
			}

			my, err := New(conf)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, my.ConnStr)
		})
	}
}

func TestMySQLIssue673(t *testing.T) {
	// Specific test case for issue #673: MySQL Native Client password parser error
	// This tests the exact scenario described in the GitHub issue
	conf := conf.Options{
		Host:       "localhost:3306",
		DriverType: conf.MySQL,
		Username:   "root",
		Password:   "AB10$CCC123", // Password with dollar sign from the issue
		Data: map[string]string{
			"test:product:name:id:1": "EaseProbe",
			"test:employee:age:id:2": "45",
		},
	}

	my, err := New(conf)
	assert.Nil(t, err)
	assert.NotNil(t, my)

	// Verify the connection string has URL-encoded password
	expectedConnStr := "root:AB10%24CCC123@tcp(localhost:3306)/?timeout=0s"
	assert.Equal(t, expectedConnStr, my.ConnStr)

	// Verify that the MySQL client was created successfully
	assert.Equal(t, "MySQL", my.Kind())
	assert.Equal(t, conf.Username, my.Username)
	assert.Equal(t, conf.Password, my.Password)
	assert.Equal(t, conf.Host, my.Host)
}

func TestData(t *testing.T) {
	monkey.Patch(sql.Open, func(driverName, dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	})
	var db *sql.DB
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Close", func(_ *sql.DB) error {
		return nil
	})

	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.MySQL,
		Username:   "username",
		Password:   "password",
		Data: map[string]string{
			"": "",
		},
	}
	my, err := New(conf)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Empty SQL data")

	conf.Data = map[string]string{
		"key": "value",
	}
	my, err = New(conf)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Invalid SQL data")

	conf.Data = map[string]string{
		"database:table:column:key:value": "expected",
	}
	my, err = New(conf)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "the value must be int")

	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Ping", func(_ *sql.DB) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Query", func(_ *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
		return &sql.Rows{}, nil
	})
	var r *sql.Rows
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Close", func(_ *sql.Rows) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Next", func(_ *sql.Rows) bool {
		return true
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Scan", func(_ *sql.Rows, args ...interface{}) error {
		v := args[0].(*string)
		*v = "expected"
		return nil
	})

	conf.Data = map[string]string{
		"database:table:column:key:1": "expected",
	}
	my, err = New(conf)
	s, m := my.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	//mismatch
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Scan", func(_ *sql.Rows, args ...interface{}) error {
		v := args[0].(*string)
		*v = "unexpected"
		return nil
	})
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "alue not match")

	// scan error
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Scan", func(_ *sql.Rows, args ...interface{}) error {
		return fmt.Errorf("scan error")
	})
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "scan error")

	// Next error
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Next", func(_ *sql.Rows) bool {
		return false
	})
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "No data")

	// Query error
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Query", func(_ *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
		return nil, fmt.Errorf("query error")
	})
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "query error")

	my.Data = map[string]string{
		"key": "value",
	}
	s, m = my.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "Invalid SQL data")

	monkey.UnpatchAll()

}
