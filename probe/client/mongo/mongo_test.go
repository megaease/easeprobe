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

	mg, err := New(conf)
	assert.Nil(t, mg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "TLS Config Error")

	conf.TLS = global.TLS{}
	mg, err = New(conf)
	assert.Equal(t, "Mongo", mg.Kind())
	assert.Nil(t, err)
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
	mg, err = New(conf)
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

	mg, err = New(conf)
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

func TestDta(t *testing.T) {

	monkey.Patch(mongo.Connect, func(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
		return &mongo.Client{}, nil
	})
	var client *mongo.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Disconnect", func(_ *mongo.Client, _ context.Context) error {
		return nil
	})

	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.Mongo,
		Username:   "username",
		Password:   "password",
		Data: map[string]string{
			"": "",
		},
	}

	mg, err := New(conf)
	assert.Nil(t, mg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Database Collection name is empty")

	conf.Data = map[string]string{
		"key": "value",
	}
	mg, err = New(conf)
	assert.Nil(t, mg)
	assert.Contains(t, err.Error(), "Invalid Format")

	conf.Data = map[string]string{
		"database:collection": "{'key' : 'value'}",
	}
	mg, err = New(conf)
	assert.Nil(t, mg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid JSON input")

	var collection *mongo.Collection
	monkey.PatchInstanceMethod(reflect.TypeOf(collection), "FindOne", func(_ *mongo.Collection, _ context.Context, _ interface{}, _ ...*options.FindOneOptions) *mongo.SingleResult {
		return &mongo.SingleResult{}
	})
	var result *mongo.SingleResult
	monkey.PatchInstanceMethod(reflect.TypeOf(result), "Err", func(_ *mongo.SingleResult) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(result), "Decode", func(_ *mongo.SingleResult, _ interface{}) error {
		return nil
	})

	conf.Data = map[string]string{
		"database:collection": "{\"key\" : \"value\"}",
	}
	mg, err = New(conf)
	s, m := mg.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	monkey.UnpatchInstanceMethod(reflect.TypeOf(result), "Err")
	s, m = mg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "Error")

	mg.Data = map[string]string{
		"database:collection": "{'key' : 'value'}",
	}
	s, m = mg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "invalid JSON input")

	mg.Data = map[string]string{
		"key": "value",
	}
	s, m = mg.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "Invalid Format")

	monkey.UnpatchAll()
}
