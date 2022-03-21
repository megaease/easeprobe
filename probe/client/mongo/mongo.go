package mongo

import (
	"context"
	"fmt"

	"github.com/megaease/easeprobe/probe/client/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Kind is the type of driver
const Kind string = "Mongo"

// Mongo is the Mongo client
type Mongo struct {
	conf.Options `yaml:",inline"`
	ConnStr      string          `yaml:"conn_str"`
	Context      context.Context `yaml:"-"`
}

// New create a Redis client
func New(opt conf.Options) Mongo {
	var conn string
	if len(opt.Password) > 0 {
		conn = fmt.Sprintf("mongodb://%s:%s@%s/?connectTimeoutMS=%d",
			opt.Username, opt.Password, opt.Host, opt.Timeout.Milliseconds())
	} else {
		conn = fmt.Sprintf("mongodb://%s/?connectTimeoutMS=%d",
			opt.Host, opt.Timeout.Milliseconds())
	}


	return Mongo{
		Options: opt,
		ConnStr: conn,
		Context: context.Background(),
	}
}

// Kind return the name of client
func (r Mongo) Kind() string {
	return Kind
}

// Probe do the health check
func (r Mongo) Probe() (bool, string) {

	opt := options.Client().ApplyURI(r.ConnStr)
	opt.ServerSelectionTimeout = &r.Timeout
	opt.SetConnectTimeout(r.Timeout)

	ctx , cancel := context.WithTimeout(r.Context, r.Timeout)
	defer cancel()

	db, err := mongo.Connect(ctx, opt)
	if err != nil {
		return false, err.Error()
	}

	defer db.Disconnect(ctx)

	err = db.Ping(ctx, nil)
	if err != nil {
		return false, err.Error()
	}

	return true, "Check MongoDB Server Successfully!"

}
