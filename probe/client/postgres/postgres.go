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
	"database/sql"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Kind is the type of driver
const Kind string = "PostgreSQL"

// PostgreSQL is the PostgreSQL client
type PostgreSQL struct {
	conf.Options  `yaml:",inline"`
	ClientOptions []pgdriver.Option `yaml:"-"`
}

// New create a PostgreSQL client
func New(opt conf.Options) PostgreSQL {
	clientOptions := []pgdriver.Option{
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(opt.Host),
		pgdriver.WithUser(opt.Username),
		pgdriver.WithTimeout(opt.Timeout.Round(time.Second)),
		pgdriver.WithApplicationName(global.OrgProgVer),
	}
	if len(opt.Password) > 0 {
		clientOptions = append(clientOptions, pgdriver.WithPassword(opt.Password))
	}

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s] %s - TLS Config error - %v", Kind, opt.Name, err)
	} else {
		tls.InsecureSkipVerify = true
		clientOptions = append(clientOptions, pgdriver.WithTLSConfig(tls))
	}

	return PostgreSQL{
		Options:       opt,
		ClientOptions: clientOptions,
	}
}

// Kind return the name of client
func (r PostgreSQL) Kind() string {
	return Kind
}

// Probe do the health check
func (r PostgreSQL) Probe() (bool, string) {
	db := sql.OpenDB(pgdriver.NewConnector(r.ClientOptions...))
	defer db.Close()

	if err := db.Ping(); err != nil {
		return false, err.Error()
	}

	// run a SQL to test
	if _, err := db.Query(`SELECT 1`); err != nil {
		return false, err.Error()
	}

	return true, "Check PostgreSQL Server Successfully!"
}
