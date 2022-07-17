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
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of driver
const Kind string = "MySQL"

// MySQL is the MySQL client
type MySQL struct {
	conf.Options `yaml:",inline"`
	tls          *tls.Config `yaml:"-"`
	ConnStr      string      `yaml:"conn_str"`
}

// New create a Mysql client
func New(opt conf.Options) MySQL {

	var conn string
	if len(opt.Password) > 0 {
		conn = fmt.Sprintf("%s:%s@tcp(%s)/?timeout=%s",
			opt.Username, opt.Password, opt.Host, opt.Timeout().Round(time.Second))
	} else {
		conn = fmt.Sprintf("%s@tcp(%s)/?timeout=%s",
			opt.Username, opt.Host, opt.Timeout().Round(time.Second))
	}

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s / %s / %s] - TLS Config error - %v", opt.ProbeKind, opt.ProbeName, opt.ProbeTag, err)
	} else if tls != nil {
		conn += "&tls=" + global.DefaultProg
	}

	return MySQL{
		Options: opt,
		tls:     tls,
		ConnStr: conn,
	}
}

// Kind return the name of client
func (r MySQL) Kind() string {
	return Kind
}

// Probe do the health check
func (r MySQL) Probe() (bool, string) {

	if r.tls != nil {
		mysql.RegisterTLSConfig(global.DefaultProg, r.tls)
	}

	db, err := sql.Open("mysql", r.ConnStr)
	if err != nil {
		return false, err.Error()
	}
	defer db.Close()

	// Check if we need to query specific data
	if len(r.Data) > 0 {
		for k, v := range r.Data {
			log.Debugf("[%s / %s / %s] - Verifying Data - [%s] : [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, k, v)
			sql, err := r.getSQL(k)
			if err != nil {
				return false, err.Error()
			}
			rows, err := db.Query(sql)
			if err != nil {
				return false, err.Error()
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
		}
	} else {
		err = db.Ping()
		if err != nil {
			return false, err.Error()
		}
		row, err := db.Query("show status like \"uptime\"") // run a SQL to test
		if err != nil {
			return false, err.Error()
		}
		defer row.Close()
	}

	return true, "Check MySQL Server Successfully!"

}

func (r MySQL) getSQL(str string) (string, error) {
	if len(strings.TrimSpace(str)) == 0 {
		return "", fmt.Errorf("Empty SQL data")
	}
	fields := strings.Split(str, ":")
	if len(fields) != 5 {
		return "", fmt.Errorf("Invalid SQL data - [%s]. (syntax: database:table:field:key:value)", str)
	}
	db := fields[0]
	table := fields[1]
	field := fields[2]
	key := fields[3]
	value := fields[4]
	//check value is int or not
	if _, err := strconv.Atoi(value); err != nil {
		return "", fmt.Errorf("Invalid SQL data - [%s], the value must be int", str)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s.%s WHERE %s = %s", field, db, table, key, value)
	return sql, nil
}
