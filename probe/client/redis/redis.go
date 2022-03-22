package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/megaease/easeprobe/probe/client/conf"
)

// Kind is the type of driver
const Kind string = "Redis"

// Redis is the Redis client
type Redis struct {
	conf.Options `yaml:",inline"`
	Context      context.Context `yaml:"-"`
}

// New create a Redis client
func New(opt conf.Options) Redis {
	return Redis{
		Options: opt,
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
