package mysql

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
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
