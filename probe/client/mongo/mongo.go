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

// Package mongo implements a probe client for the MongoDB database.
package mongo

import (
	"context"
	"fmt"
	"strings"

	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Kind is the type of driver
const Kind string = "Mongo"

// Mongo is the Mongo client
type Mongo struct {
	conf.Options `yaml:",inline"`
	ConnStr      string                 `yaml:"conn_str,omitempty" json:"conn_str,omitempty"`
	ClientOpt    *options.ClientOptions `yaml:"-" json:"-"`
	Context      context.Context        `yaml:"-" json:"-"`
}

// New create a Mongo client
func New(opt conf.Options) (*Mongo, error) {
	var conn string
	if len(opt.Password) > 0 {
		conn = fmt.Sprintf("mongodb://%s:%s@%s/?connectTimeoutMS=%d",
			opt.Username, opt.Password, opt.Host, opt.Timeout().Milliseconds())
	} else {
		conn = fmt.Sprintf("mongodb://%s/?connectTimeoutMS=%d",
			opt.Host, opt.Timeout().Milliseconds())
	}

	log.Debugf("[%s / %s / %s] - Connection - %s", opt.ProbeKind, opt.ProbeName, opt.ProbeTag, conn)

	var maxConn uint64 = 1
	client := options.Client().ApplyURI(conn)
	client.ServerSelectionTimeout = &opt.ProbeTimeout
	client.ConnectTimeout = &opt.ProbeTimeout
	client.MaxConnecting = &maxConn
	client.MaxPoolSize = &maxConn
	client.MinPoolSize = &maxConn

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s / %s / %s] - TLS Config Error - %v", opt.ProbeKind, opt.ProbeName, opt.ProbeTag, err)
		return nil, fmt.Errorf("TLS Config Error - %v", err)
	} else if tls != nil {
		client.TLSConfig = tls
		client.SetAuth(options.Credential{AuthMechanism: "MONGODB-X509"})
	}

	mongo := &Mongo{
		Options:   opt,
		ConnStr:   conn,
		ClientOpt: client,
		Context:   context.Background(),
	}

	if err := mongo.checkData(); err != nil {
		return nil, err
	}

	return mongo, nil
}

// Kind return the name of client
func (r *Mongo) Kind() string {
	return Kind
}

// checkData do the data checking
func (r *Mongo) checkData() error {

	for k, v := range r.Data {
		if _, _, err := getDBCollection(k); err != nil {
			return err
		}
		var bdoc interface{}
		if err := bson.UnmarshalExtJSON([]byte(v), true, &bdoc); err != nil {
			return err
		}
	}

	return nil
}

// Probe do the health check
func (r *Mongo) Probe() (bool, string) {

	ctx, cancel := context.WithTimeout(r.Context, r.Timeout())
	defer cancel()

	db, err := mongo.Connect(ctx, r.ClientOpt)
	if err != nil {
		return false, err.Error()
	}

	defer db.Disconnect(ctx)

	if len(r.Data) > 0 {
		for key, value := range r.Data {
			log.Debugf("[%s / %s / %s] - Verifying Data - [%s]: [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, key, value)
			dbName, collectionName, err := getDBCollection(key)
			if err != nil {
				return false, fmt.Sprintf("[%s] Error - %v", key, err)
			}
			collection := db.Database(dbName).Collection(collectionName)
			var bdoc interface{}
			err = bson.UnmarshalExtJSON([]byte(value), true, &bdoc)
			if err != nil {
				return false, fmt.Sprintf("[%s] Error - %v", value, err)
			}
			result := collection.FindOne(ctx, bdoc)
			if err := result.Err(); err != nil {
				return false, fmt.Sprintf("Find [%s] Error - %v", value, err)
			}
			var doc bson.M
			result.Decode(&doc)
			log.Debugf("[%s / %s / %s] - Find [%s] - %+v", r.ProbeKind, r.ProbeName, r.ProbeTag, value, doc)
			log.Debugf("[%s / %s / %s] - Data Verified Successfully - [%s]: [%s]", r.ProbeKind, r.ProbeName, r.ProbeTag, key, value)
		}
	} else {
		// Call Ping to verify that the deployment is up and the Client was
		// configured successfully. As mentioned in the Ping documentation, this
		// reduces application resiliency as the server may be temporarily
		// unavailable when Ping is called.
		err = db.Ping(ctx, readpref.Primary())
		if err != nil {
			return false, err.Error()
		}
	}

	return true, "Check MongoDB Server Successfully!"

}

func getDBCollection(str string) (database, collection string, err error) {
	if len(strings.TrimSpace(str)) == 0 {
		return "", "", fmt.Errorf("Database Collection name is empty")
	}
	fields := strings.Split(str, ":")
	if len(fields) != 2 {
		err = fmt.Errorf("Invalid Format - [%s] (syntax: database.collection) ", str)
		return
	}
	database = fields[0]
	collection = fields[1]
	return
}
