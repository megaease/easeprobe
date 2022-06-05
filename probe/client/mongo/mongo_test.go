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

package mongo

import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestMongo(t *testing.T) {
	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.Mongo,
		Username:   "username",
		Password:   "password",
		TLS: global.TLS{
			CA:   "ca",
			Cert: "cert",
			Key:  "key",
		},
	}

	mg := New(conf)
	assert.Equal(t, "Mongo", mg.Kind())
	connStr := fmt.Sprintf("mongodb://%s:%s@%s/?connectTimeoutMS=%d",
		conf.Username, conf.Password, conf.Host, conf.Timeout().Milliseconds())
	assert.Equal(t, connStr, mg.ConnStr)
	assert.Nil(t, mg.ClientOpt.TLSConfig)

	monkey.Patch(mongo.Connect, func(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
		return &mongo.Client{}, nil
	})
	var client *mongo.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Disconnect", func(_ *mongo.Client, _ context.Context) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Ping", func(_ *mongo.Client, ctx context.Context, rp *readpref.ReadPref) error {
		return nil
	})

	s, m := mg.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	conf.Password = ""
	mg = New(conf)
	connStr = fmt.Sprintf("mongodb://%s/?connectTimeoutMS=%d",
		conf.Host, conf.Timeout().Milliseconds())
	assert.Equal(t, connStr, mg.ConnStr)

	s, m = mg.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	var tc *global.TLS
	monkey.PatchInstanceMethod(reflect.TypeOf(tc), "Config", func(_ *global.TLS) (*tls.Config, error) {
		return &tls.Config{}, nil
	})

	mg = New(conf)
	assert.Equal(t, "Mongo", mg.Kind())
	assert.Equal(t, connStr, mg.ConnStr)
	assert.NotNil(t, mg.ClientOpt.TLSConfig)

	s, m = mg.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	//Ping Error
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Ping", func(_ *mongo.Client, ctx context.Context, rp *readpref.ReadPref) error {
		return fmt.Errorf("ping error")
	})
	s, m = mg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "ping error")

	//Connect Error
	monkey.Patch(mongo.Connect, func(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
		return nil, fmt.Errorf("connect error")
	})
	s, m = mg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "connect error")

	monkey.UnpatchAll()
}
