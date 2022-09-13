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

// Package redis is the native client probe for Redis
package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of driver
const Kind string = "Redis"

// Redis is the Redis client
type Redis struct {
	conf.Options `yaml:",inline"`
	tls          *tls.Config     `yaml:"-" json:"-"`
	Context      context.Context `yaml:"-" json:"-"`
}

// New create a Redis client
func New(opt conf.Options) (*Redis, error) {

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s / %s / %s] - TLS Config Error - %v", opt.ProbeKind, opt.ProbeName, opt.ProbeTag, err)
		return nil, fmt.Errorf("TLS Config Error - %v", err)
	}

	r := Redis{
		Options: opt,
		tls:     tls,
		Context: context.Background(),
	}
	return &r, nil
}

// Kind return the name of client
func (r *Redis) Kind() string {
	return Kind
}

// Probe do the health check
func (r *Redis) Probe() (bool, string) {

	rdb := redis.NewClient(&redis.Options{
		Addr:        r.Host,
		Password:    r.Password,  // no password set
		DB:          0,           // use default DB
		DialTimeout: r.Timeout(), // dial timout
		TLSConfig:   r.tls,       //tls
	})

	ctx, cancel := context.WithTimeout(r.Context, r.Timeout())
	defer cancel()
	defer rdb.Close()

	// Check if we need to query specific keys or not
	if len(r.Data) > 0 {
		for k, v := range r.Data {
			log.Debugf("[%s / %s / %s] Verifying Data -  key = [%s], value = [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, k, v)
			val, err := rdb.Get(ctx, k).Result()
			if err != nil {
				return false, fmt.Sprintf("Get Key [%s] Error - %v", k, err)
			}
			if val != v {
				return false, fmt.Sprintf("Key [%s] expected [%s] got [%s]", k, v, val)
			}
			log.Debugf("[%s / %s / %s] Data Verified Successfully! key= [%s], value = [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, k, v)
		}
	} else {
		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			return false, err.Error()
		}
	}
	return true, "Ping Redis Server Successfully!"
}
