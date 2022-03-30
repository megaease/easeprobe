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

// New create a Redis client
func New(opt conf.Options) MySQL {

	var conn string
	if len(opt.Password) > 0 {
		conn = fmt.Sprintf("%s:%s@tcp(%s)/?timeout=%s",
			opt.Username, opt.Password, opt.Host, opt.Timeout.Round(time.Second))
	} else {
		conn = fmt.Sprintf("%s@tcp(%s)/?timeout=%s",
			opt.Username, opt.Host, opt.Timeout.Round(time.Second))
	}

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s] %s - TLS Config error - %v", Kind, opt.Name, err)
	} else {
		conn += "&tls=" + global.Prog
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
		mysql.RegisterTLSConfig(global.Prog, r.tls)
	}

	db, err := sql.Open("mysql", r.ConnStr)
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return false, err.Error()
	}
	_, err = db.Query("show status like \"uptime\"") // run a SQL to test
	if err != nil {
		return false, err.Error()
	}

	return true, "Check MySQL Server Successfully!"

}
