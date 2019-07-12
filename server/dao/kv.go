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

//Package dao is a persis layer of kie
package dao

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"time"
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

//Options mongodb options
type Options struct {
	URI      string
	PoolSize int
	SSL      bool
	TLS      *tls.Config
	Timeout  time.Duration
}

//NewKVService create a kv service
//TODO, multiple config server
func NewKVService() (*MongodbService, error) {
	var d time.Duration
	var err error
	if config.GetDB().Timeout != "" {
		d, err = time.ParseDuration(config.GetDB().Timeout)
		if err != nil {
			return nil, errors.New("timeout setting invalid:" + config.GetDB().Timeout)
		}
	}
	opts := Options{
		URI:      config.GetDB().URI,
		PoolSize: config.GetDB().PoolSize,
		SSL:      config.GetDB().SSLEnabled,
		Timeout:  d,
	}
	if opts.SSL {
		if config.GetDB().RootCA == "" {
			return nil, ErrRootCAMissing
		}
		pool := x509.NewCertPool()
		caCert, err := ioutil.ReadFile(config.GetDB().RootCA)
		if err != nil {
			return nil, fmt.Errorf("read ca cert file %s failed", caCert)
		}
		pool.AppendCertsFromPEM(caCert)
		opts.TLS = &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: true,
		}

	}
	return NewMongoService(opts)
}
func (s *MongodbService) findOneKey(ctx context.Context, filter bson.M) ([]*model.KVDoc, error) {
	collection := s.c.Database(DB).Collection(CollectionKV)
	sr := collection.FindOne(ctx, filter)
	if sr.Err() != nil {
		return nil, sr.Err()
	}
	curKV := &model.KVDoc{}
	err := sr.Decode(curKV)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrKeyNotExists
		}
		openlogging.Error("decode error: " + err.Error())
		return nil, err
	}
	return []*model.KVDoc{curKV}, nil
}

//KVExist supports you query by label map or labels id
func (s *MongodbService) KVExist(ctx context.Context, domain, key string, options ...FindOption) (primitive.ObjectID, error) {
	opts := FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.LabelID != "" {
		kvs, err := s.FindKVByLabelID(ctx, domain, opts.LabelID, key)
		if err != nil {
			return primitive.NilObjectID, err
		}
		return kvs[0].ID, nil
	}
	kvs, err := s.FindKV(ctx, domain, WithExactLabels(), WithLabels(opts.Labels), WithKey(key))
	if err != nil {
		return primitive.NilObjectID, err
	}
	if len(kvs) != 1 {
		return primitive.NilObjectID, ErrTooMany
	}

	return kvs[0].Data[0].ID, nil

}

func (s *MongodbService) findKV(ctx context.Context, domain string, opts FindOptions) (*mongo.Cursor, error) {
	collection := s.c.Database(DB).Collection(CollectionKV)
	ctx, _ = context.WithTimeout(ctx, s.timeout)
	filter := bson.M{"domain": domain}
	if opts.Key != "" {
		filter["key"] = opts.Key
	}
	for k, v := range opts.Labels {
		filter["labels."+k] = v
	}

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find kv failed, deadline exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": s.timeout,
			}))
			return nil, fmt.Errorf("can not reach mongodb in %s", s.timeout)
		}
		return nil, err
	}
	return cur, err
}

//DeleteKV by kvID
func (s *MongodbService) DeleteKV(ctx context.Context, hexID primitive.ObjectID) error {
	collection := s.c.Database(DB).Collection(CollectionKV)
	dr, err := collection.DeleteOne(ctx, bson.M{"_id": hexID})
	//check error and delete number
	if err != nil {
		openlogging.Error(fmt.Sprintf("delete [%s] failed : [%s]", hexID, err))
		return err
	}
	if dr.DeletedCount != 1 {
		openlogging.Warn(fmt.Sprintf("Failed,May have been deleted,kvID=%s", hexID))
	} else {
		openlogging.Info(fmt.Sprintf("delete success,kvID=%s", hexID))
	}
	return err
}
