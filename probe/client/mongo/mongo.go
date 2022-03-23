package mongo

import (
	"context"
	"fmt"

	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Kind is the type of driver
const Kind string = "Mongo"

// Mongo is the Mongo client
type Mongo struct {
	conf.Options `yaml:",inline"`
	ConnStr      string                 `yaml:"conn_str"`
	ClientOpt    *options.ClientOptions `yaml:"-"`
	Context      context.Context        `yaml:"-"`
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

	var maxConn uint64 = 1
	client := options.Client().ApplyURI(conn)
	client.ServerSelectionTimeout = &opt.Timeout
	client.ConnectTimeout = &opt.Timeout
	client.MaxConnecting = &maxConn
	client.MaxPoolSize = &maxConn
	client.MinPoolSize = &maxConn

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s] %s - TLS Config error - %v", Kind, opt.Name, err)
	} else {
		client.TLSConfig = tls
		client.SetAuth(options.Credential{AuthMechanism: "MONGODB-X509"})
	}

	return Mongo{
		Options:   opt,
		ConnStr:   conn,
		ClientOpt: client,
		Context:   context.Background(),
	}
}

// Kind return the name of client
func (r Mongo) Kind() string {
	return Kind
}

// Probe do the health check
func (r Mongo) Probe() (bool, string) {

	ctx, cancel := context.WithTimeout(r.Context, r.Timeout)
	defer cancel()

	log.Debugln(r.ClientOpt)

	db, err := mongo.Connect(ctx, r.ClientOpt)
	if err != nil {
		return false, err.Error()
	}

	defer db.Disconnect(ctx)

	// Call Ping to verify that the deployment is up and the Client was
	// configured successfully. As mentioned in the Ping documentation, this
	// reduces application resiliency as the server may be temporarily
	// unavailable when Ping is called.
	err = db.Ping(ctx, readpref.Primary())
	if err != nil {
		return false, err.Error()
	}

	return true, "Check MongoDB Server Successfully!"

}
