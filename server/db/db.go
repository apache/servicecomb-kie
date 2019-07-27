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

//Package db manage db connection
package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"time"
)

//const for db name and collection name
const (
	Name                    = "kie"
	CollectionLabel         = "label"
	CollectionKV            = "kv"
	CollectionLabelRevision = "label_revision"

	DefaultTimeout   = 5 * time.Second
	DefaultValueType = "text"
)

//db errors
var (
	ErrMissingDomain          = errors.New("domain info missing, illegal access")
	ErrKeyNotExists           = errors.New("key with labels does not exits")
	ErrLabelNotExists         = errors.New("labels does not exits")
	ErrTooMany                = errors.New("key with labels should be only one")
	ErrKeyMustNotEmpty        = errors.New("must supply key if you want to get exact one result")
	ErrRevisionNotExist       = errors.New("label revision not exist")
	ErrKVIDIsNil              = errors.New("kvID id is nil")
	ErrKvIDAndLabelIDNotMatch = errors.New("kvID and labelID do not match")
	ErrRootCAMissing          = errors.New("rootCAFile is empty in config file")
)

var client *mongo.Client

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
	return nil
}

//GetClient create a new mongo db client
//if client is created, just return.
func GetClient() (*mongo.Client, error) {
	if client == nil {
		var err error
		clientOps := []*options.ClientOptions{options.Client().ApplyURI(config.GetDB().URI)}
		if config.GetDB().SSLEnabled {
			if config.GetDB().RootCA == "" {
				return nil, ErrRootCAMissing
			}
			pool := x509.NewCertPool()
			caCert, err := ioutil.ReadFile(config.GetDB().RootCA)
			if err != nil {
				return nil, fmt.Errorf("read ca cert file %s failed", caCert)
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
			return nil, err
		}
		openlogging.Info("DB connecting")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Connect(ctx)
		if err != nil {
			return nil, err
		}
		openlogging.Info("DB connected")
	}
	return client, nil
}
