/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//Package session manage db connection
package session

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"reflect"
	"sync"
	"time"

	"github.com/apache/servicecomb-kie/server/config"
	"go.mongodb.org/mongo-driver/mongo"
)

//const for db name and collection name
const (
	DBName = "kie"

	CollectionLabel      = "label"
	CollectionKV         = "kv"
	CollectionKVRevision = "kv_revision"
	CollectionCounter    = "counter"
	CollectionView       = "view"
	DefaultTimeout       = 5 * time.Second
	DefaultValueType     = "text"
)

//db errors
var (
	ErrMissingDomain   = errors.New("domain info missing, illegal access")
	ErrMissingProject  = errors.New("project info missing, illegal access")
	ErrLabelNotExists  = errors.New("labels does not exits")
	ErrTooMany         = errors.New("key with labels should be only one")
	ErrKeyMustNotEmpty = errors.New("must supply key if you want to get exact one result")

	ErrIDIsNil                = errors.New("id is empty")
	ErrKvIDAndLabelIDNotMatch = errors.New("kvID and labelID do not match")
	ErrRootCAMissing          = errors.New("rootCAFile is empty in config file")

	ErrViewCreation = errors.New("can not create view")
	ErrViewUpdate   = errors.New("can not update view")
	ErrViewNotExist = errors.New("view not exists")
	ErrViewFinding  = errors.New("view search error")
)

var client *mongo.Client
var once sync.Once
var db *mongo.Database

//Timeout db operation time out
var Timeout time.Duration

//Init prepare params
func Init() error {
	var err error
	if config.GetDB().Timeout != "" {
		Timeout, err = time.ParseDuration(config.GetDB().Timeout)
		if err != nil {
			return errors.New("timeout setting invalid:" + config.GetDB().Timeout)
		}
	}
	if Timeout == 0 {
		Timeout = DefaultTimeout
	}
	once.Do(func() {
		sc, _ := bsoncodec.NewStructCodec(bsoncodec.DefaultStructTagParser)
		reg := bson.NewRegistryBuilder().
			RegisterEncoder(reflect.TypeOf(model.LabelDoc{}), sc).
			RegisterEncoder(reflect.TypeOf(model.KVDoc{}), sc).
			Build()
		clientOps := []*options.ClientOptions{options.Client().ApplyURI(config.GetDB().URI)}
		if config.GetDB().SSLEnabled {
			if config.GetDB().RootCA == "" {
				err = ErrRootCAMissing
				return
			}
			pool := x509.NewCertPool()
			caCert, err := ioutil.ReadFile(config.GetDB().RootCA)
			if err != nil {
				err = fmt.Errorf("read ca cert file %s failed", caCert)
				return
			}
			pool.AppendCertsFromPEM(caCert)
			tc := &tls.Config{
				RootCAs:            pool,
				InsecureSkipVerify: true,
			}
			clientOps = append(clientOps, options.Client().SetTLSConfig(tc))
			openlogging.Info("enabled ssl communication to mongodb")
		}
		client, err = mongo.NewClient(clientOps...)
		if err != nil {
			return
		}
		openlogging.Info("DB connecting")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Connect(ctx)
		if err != nil {
			return
		}
		openlogging.Info("DB connected")
		db = client.Database(DBName, &options.DatabaseOptions{
			Registry: reg,
		})

	})
	return nil
}

//GetDB get mongo db client
func GetDB() *mongo.Database {
	return db
}

func CreateView(ctx context.Context, view, source string, pipeline mongo.Pipeline) error {
	sr := GetDB().RunCommand(ctx,
		bson.D{
			{"create", view},
			{"viewOn", source},
			{"pipeline", pipeline},
		})
	if sr.Err() != nil {
		openlogging.Error("can not create view: " + sr.Err().Error())
		return ErrViewCreation
	}
	return nil
}
