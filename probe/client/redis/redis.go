package redis

import (
	"context"
	"crypto/tls"

	"github.com/go-redis/redis/v8"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of driver
const Kind string = "Redis"

// Redis is the Redis client
type Redis struct {
	conf.Options `yaml:",inline"`
	tls          *tls.Config     `yaml:"-"`
	Context      context.Context `yaml:"-"`
}

// New create a Redis client
func New(opt conf.Options) Redis {

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s] %s - TLS Config error - %v", Kind, opt.Name, err)
	}

	return Redis{
		Options: opt,
		tls:     tls,
		Context: context.Background(),
	}
}

// Kind return the name of client
func (r Redis) Kind() string {
	return Kind
}

// Probe do the health check
func (r Redis) Probe() (bool, string) {

	rdb := redis.NewClient(&redis.Options{
		Addr:        r.Host,
		Password:    r.Password, // no password set
		DB:          0,          // use default DB
		DialTimeout: r.Timeout,  // dial timout
		TLSConfig:   r.tls,      //tls
	})

	ctx, cancel := context.WithTimeout(r.Context, r.Timeout)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()

	defer rdb.Close()

	if err != nil {
		return false, err.Error()
	}
	return true, "Ping Redis Server Successfully!"

}
