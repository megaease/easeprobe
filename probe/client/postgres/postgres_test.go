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

package postgres

import (
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun/driver/pgdriver"
)

func TestPostgreSQL(t *testing.T) {
	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.PostgreSQL,
		Username:   "username",
		Password:   "password",
		TLS: global.TLS{
			CA:   "ca",
			Cert: "cert",
			Key:  "key",
		},
	}

	pg, err := New(conf)
	assert.Nil(t, pg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "TLS Config Error")

	conf.TLS = global.TLS{}
	pg, err = New(conf)
	assert.Nil(t, err)
	assert.Equal(t, "PostgreSQL", pg.Kind())
	pgd := pgdriver.NewConnector(pg.ClientOptions...)
	assert.Equal(t, conf.Host, pgd.Config().Addr)
	assert.Equal(t, conf.Username, pgd.Config().User)
	assert.Equal(t, conf.Password, pgd.Config().Password)
	assert.Equal(t, conf.Timeout(), pgd.Config().DialTimeout)
	assert.Equal(t, conf.Timeout(), pgd.Config().ReadTimeout)
	assert.Equal(t, conf.Timeout(), pgd.Config().WriteTimeout)
	assert.Nil(t, pgd.Config().TLSConfig)

	monkey.Patch(sql.OpenDB, func(c driver.Connector) *sql.DB {
		return &sql.DB{}
	})

	var db *sql.DB
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Ping", func(_ *sql.DB) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Close", func(_ *sql.DB) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Query", func(_ *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
		return &sql.Rows{}, nil
	})
	var r *sql.Rows
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Close", func(_ *sql.Rows) error {
		return nil
	})

	s, m := pg.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	s, m = pg.ProbeWithDataChecking()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	// TLS config success
	var tc *global.TLS
	monkey.PatchInstanceMethod(reflect.TypeOf(tc), "Config", func(_ *global.TLS) (*tls.Config, error) {
		return &tls.Config{}, nil
	})

	pg, err = New(conf)
	pgd = pgdriver.NewConnector(pg.ClientOptions...)
	assert.True(t, pgd.Config().TLSConfig.InsecureSkipVerify)

	// Query error
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Query", func(_ *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
		return nil, fmt.Errorf("query error")
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "query error")

	// Ping error
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Ping", func(_ *sql.DB) error {
		return fmt.Errorf("ping error")
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "ping error")

	// OpenDB error
	monkey.Patch(sql.OpenDB, func(c driver.Connector) *sql.DB {
		return nil
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "OpenDB error")

	monkey.UnpatchAll()

}

func TestData(t *testing.T) {
	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.PostgreSQL,
		Username:   "username",
		Password:   "password",
		Data: map[string]string{
			"": "",
		},
	}

	pg, err := New(conf)
	assert.Nil(t, pg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Empty SQL data")

	conf.Data = map[string]string{
		"sql": "",
	}
	pg, err = New(conf)
	assert.Nil(t, pg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Invalid SQL data")

	conf.Data = map[string]string{
		"database:table:column:key:value": "excepted",
	}
	pg, err = New(conf)
	assert.Nil(t, pg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "the value must be int")

	monkey.Patch(sql.OpenDB, func(c driver.Connector) *sql.DB {
		return &sql.DB{}
	})
	var db *sql.DB
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Close", func(_ *sql.DB) error {
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
	pg, err = New(conf)
	s, m := pg.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Scan", func(_ *sql.Rows, args ...interface{}) error {
		v := args[0].(*string)
		*v = "unexpected"
		return nil
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "Value not match")

	// scan error
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Scan", func(_ *sql.Rows, args ...interface{}) error {
		return fmt.Errorf("scan error")
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "scan error")

	// next error
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "Next", func(_ *sql.Rows) bool {
		return false
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "No data")

	// query error
	monkey.PatchInstanceMethod(reflect.TypeOf(db), "Query", func(_ *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
		return nil, fmt.Errorf("query error")
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "query error")

	// OpenDB error
	monkey.Patch(sql.OpenDB, func(c driver.Connector) *sql.DB {
		return nil
	})
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "OpenDB error")

	pg.Data = map[string]string{
		"key": "value",
	}
	s, m = pg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "Invalid SQL data")

	monkey.UnpatchAll()
}
