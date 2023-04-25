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

// Package postgres is the native client probe for  PostgreSQL
package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Kind is the type of driver
const Kind string = "PostgreSQL"

// revive:disable
// PostgreSQL is the PostgreSQL client
type PostgreSQL struct {
	conf.Options  `yaml:",inline"`
	ClientOptions []pgdriver.Option `yaml:"-" json:"-"`
}

// revive:enable

// New create a PostgreSQL client
func New(opt conf.Options) (*PostgreSQL, error) {
	clientOptions := []pgdriver.Option{
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(opt.Host),
		pgdriver.WithUser(opt.Username),
		pgdriver.WithTimeout(opt.Timeout().Round(time.Second)),
		pgdriver.WithApplicationName(global.OrgProgVer),
	}
	if len(opt.Password) > 0 {
		clientOptions = append(clientOptions, pgdriver.WithPassword(opt.Password))
	}

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s / %s / %s] - TLS Config Error - %v", opt.ProbeKind, opt.ProbeName, opt.ProbeTag, err)
		return nil, fmt.Errorf("TLS Config Error - %v", err)
	} else if tls != nil {
		tls.InsecureSkipVerify = true
	}
	// if the tls is nil which means `sslmode=disable`
	clientOptions = append(clientOptions, pgdriver.WithTLSConfig(tls))

	pg := &PostgreSQL{
		Options:       opt,
		ClientOptions: clientOptions,
	}
	if err := pg.checkData(); err != nil {
		return nil, err
	}
	return pg, nil
}

// Kind return the name of client
func (r *PostgreSQL) Kind() string {
	return Kind
}

// checkData do the data checking
func (r *PostgreSQL) checkData() error {

	for k := range r.Data {
		_, _, err := r.getSQL(k)
		if err != nil {
			return err
		}
	}

	return nil
}

// Probe do the health check
func (r *PostgreSQL) Probe() (bool, string) {

	if len(r.Data) > 0 {
		return r.ProbeWithDataChecking()
	}
	return r.ProbeWithPing()
}

// ProbeWithPing do the health check with ping & Select 1;
func (r *PostgreSQL) ProbeWithPing() (bool, string) {
	r.ClientOptions = append(r.ClientOptions, pgdriver.WithDatabase("template1"))
	db := sql.OpenDB(pgdriver.NewConnector(r.ClientOptions...))
	if db == nil {
		return false, "OpenDB error"
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return false, err.Error()
	}

	// run a SQL to test
	row, err := db.Query(`SELECT 1`)
	if err != nil {
		return false, err.Error()
	}
	row.Close()
	return true, "Check PostgreSQL Server Successfully!"
}

// ProbeWithDataChecking do the health check with data checking
func (r *PostgreSQL) ProbeWithDataChecking() (bool, string) {
	if len(r.Data) == 0 {
		log.Warnf("[%s / %s / %s] - No data found, use ping instead", r.ProbeKind, r.ProbeName, r.ProbeTag)
		return r.ProbeWithPing()
	}

	for k, v := range r.Data {
		if ok, msg := r.verifyData(k, v); !ok {
			return ok, msg
		}
	}

	return true, "Check PostgreSQL Server Successfully!"
}

func (r *PostgreSQL) verifyData(k, v string) (bool, string) {
	log.Debugf("[%s / %s / %s] - Verifying Data - [%s] : [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, k, v)
	//connect to the database
	dbName, sqlstr, err := r.getSQL(k)
	if err != nil {
		return false, fmt.Sprintf("Invalid SQL data - [%s], %v", v, err)
	}
	clientOptions := append(r.ClientOptions, pgdriver.WithDatabase(dbName))
	db := sql.OpenDB(pgdriver.NewConnector(clientOptions...))
	if db == nil {
		return false, "OpenDB error"
	}
	defer db.Close()

	// query the data
	log.Debugf("[%s / %s / %s] - SQL - [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, sqlstr)
	rows, err := db.Query(sqlstr)
	if err != nil {
		return false, fmt.Sprintf("Query error - [%s], %v", v, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return false, fmt.Sprintf("No data found for [%s]", k)
	}
	//check the value is equal to the value in data
	var value string
	if err := rows.Scan(&value); err != nil {
		return false, err.Error()
	}
	if value != v {
		return false, fmt.Sprintf("Value not match for [%s] expected [%s] got [%s] ", k, v, value)
	}

	log.Debugf("[%s / %s / %s] - Data Verified Successfully! - [%s] : [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, k, v)
	return true, "Check PostgreSQL Server Successfully!"
}

// getSQL get the SQL statement
// input: database:table:column:key:value
// output: SELECT column FROM database.table WHERE key = value
func (r *PostgreSQL) getSQL(str string) (string, string, error) {
	if len(strings.TrimSpace(str)) == 0 {
		return "", "", fmt.Errorf("Empty SQL data")
	}
	fields := strings.Split(str, ":")
	if len(fields) != 5 {
		return "", "", fmt.Errorf("Invalid SQL data - [%s]. (syntax: database:table:field:key:value)", str)
	}
	db := global.EscapeQuote(fields[0])
	table := global.EscapeQuote(fields[1])
	field := global.EscapeQuote(fields[2])
	key := global.EscapeQuote(fields[3])
	value := global.EscapeQuote(fields[4])
	//check value is int or not
	if _, err := strconv.Atoi(value); err != nil {
		return "", "", fmt.Errorf("Invalid SQL data - [%s], the value must be int", str)
	}

	sql := fmt.Sprintf(`SELECT "%s" FROM "%s" WHERE "%s" = %s`, field, table, key, value)
	return db, sql, nil
}
