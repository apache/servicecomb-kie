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

package kv

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var client *mongo.Client

const (
	DB                 = "kie"
	CollectionKV       = "kv"
	CollectionRevision = "revision"
	DefaultTimeout     = 5 * time.Second
	DefaultValueType   = "text"
)

type MongodbService struct {
	c       *mongo.Client
	timeout time.Duration
}

func (s *MongodbService) CreateOrUpdate(kv *model.KV) (*model.KV, error) {
	if kv.Domain == "" {
		return nil, ErrMissingDomain
	}
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
	collection := s.c.Database(DB).Collection(CollectionKV)
	oid, err := s.Exist(kv.Key, kv.Domain, kv.Labels)
	if err != nil {
		if err != ErrNotExists {
			return nil, err
		}
	}
	if oid != "" {
		hex, err := primitive.ObjectIDFromHex(oid)
		if err != nil {
			openlogging.Error(fmt.Sprintf("convert %s ,err:%s", oid, err))
			return nil, err
		}
		kv.ID = hex
		if err := s.update(ctx, collection, kv); err != nil {
			return nil, err
		}
		return kv, nil
	}
	if kv.ValueType == "" {
		kv.ValueType = DefaultValueType
	}
	//set 1 to revision for insertion
	kv.Revision = 1
	res, err := collection.InsertOne(ctx, kv)
	if err != nil {
		return nil, err
	}
	objectID, _ := res.InsertedID.(primitive.ObjectID)
	kv.ID = objectID
	if err := s.AddHistory(kv); err != nil {
		openlogging.Warn(
			fmt.Sprintf("can not update version for [%s] [%s] in [%s]",
				kv.Key, kv.Labels, kv.Domain))
	}
	openlogging.Debug(fmt.Sprintf("create %s with labels %s value [%s]", kv.Key, kv.Labels, kv.Value))
	return kv, nil
}

//update get latest revision from history
//and increase revision
//and update and them add new history
func (s *MongodbService) update(ctx context.Context, collection *mongo.Collection, kv *model.KV) error {
	h, err := s.getLatest(kv.ID)
	if err != nil {
		openlogging.Error(fmt.Sprintf("get latest [%s][%s] in [%s],err: %s",
			kv.Key, kv.Labels, kv.Domain, err.Error()))
		return err
	}
	if h != nil {
		kv.Revision = h.Revision + 1
	}
	ur, err := collection.UpdateOne(ctx, bson.M{"_id": kv.ID}, bson.D{
		{"$set", bson.D{
			{"value", kv.Value},
			{"revision", kv.Revision},
			{"checker", kv.Checker},
		}},
	})
	if err != nil {
		return err
	}
	openlogging.Debug(
		fmt.Sprintf("update %s with labels %s value [%s] %d ",
			kv.Key, kv.Labels, kv.Value, ur.ModifiedCount))
	if err := s.AddHistory(kv); err != nil {
		openlogging.Warn(
			fmt.Sprintf("can not update version for [%s] [%s] in [%s]",
				kv.Key, kv.Labels, kv.Domain))
	}
	openlogging.Debug(
		fmt.Sprintf("add history %s with labels %s value [%s] %d ",
			kv.Key, kv.Labels, kv.Value, ur.ModifiedCount))
	return nil

}
func (s *MongodbService) Exist(key, domain string, labels model.Labels) (string, error) {
	kvs, err := s.Find(domain, WithExactLabels(), WithLabels(labels), WithKey(key))
	if err != nil {
		return "", err
	}
	if len(kvs) != 1 {
		return "", ErrTooMany
	}

	return kvs[0].ID.Hex(), nil

}

//Find get kvs by key, labels
//because labels has a a lot of combination,
//you can use WithExactLabels to return only one kv which's labels exactly match the criteria
func (s *MongodbService) Find(domain string, options ...CallOption) ([]*model.KV, error) {
	opts := CallOptions{}
	for _, o := range options {
		o(&opts)
	}
	if domain == "" {
		return nil, ErrMissingDomain
	}
	collection := s.c.Database(DB).Collection(CollectionKV)
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
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
			return nil, ErrAction("find", opts.Key, opts.Labels, domain, fmt.Errorf("can not reach mongodb in %s", s.timeout))
		}
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	if opts.ExactLabels {
		openlogging.Debug(fmt.Sprintf("find one [%s] with lables [%s] in [%s]", opts.Key, opts.Labels, domain))
		curKV := &model.KV{} //reuse this pointer to reduce GC, only clear label
		//check label length to get the exact match
		for cur.Next(ctx) { //although complexity is O(n), but there won't be so much labels for one key
			curKV.Labels = nil
			err := cur.Decode(curKV)
			if err != nil {
				openlogging.Error("decode error: " + err.Error())
				return nil, err
			}
			if len(curKV.Labels) == len(opts.Labels) {
				openlogging.Debug("hit")
				return []*model.KV{curKV}, nil
			}

		}
		return nil, ErrNotExists
	} else {
		kvs := make([]*model.KV, 0)
		for cur.Next(ctx) {
			curKV := &model.KV{}
			if err := cur.Decode(curKV); err != nil {
				openlogging.Error("decode to KVs error: " + err.Error())
				return nil, err
			}
			kvs = append(kvs, curKV)

		}
		if len(kvs) == 0 {
			return nil, ErrNotExists
		}
		return kvs, nil
	}

}
func (s *MongodbService) DeleteByID(id string) error {
	collection := s.c.Database(DB).Collection(CollectionKV)
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		openlogging.Error(fmt.Sprintf("convert %s ,err:%s", id, err))
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
	dr, err := collection.DeleteOne(ctx, bson.M{"_id": hex})
	if err != nil {
		openlogging.Error(fmt.Sprintf("delete [%s] failed: %s", hex, err))
	}
	if dr.DeletedCount != 1 {
		openlogging.Warn(fmt.Sprintf("delete [%s], but it is not exist", hex))
	}
	return nil
}

func (s *MongodbService) Delete(key, domain string, labels model.Labels) error {
	return nil
}
func (s *MongodbService) AddHistory(kv *model.KV) error {
	collection := s.c.Database(DB).Collection(CollectionRevision)
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
	h := &model.KVHistory{
		KID:      kv.ID.Hex(),
		Value:    kv.Value,
		Revision: kv.Revision,
		Checker:  kv.Checker,
	}
	_, err := collection.InsertOne(ctx, h)
	if err != nil {
		openlogging.Error(err.Error())
		return err
	}
	return nil
}
func (s *MongodbService) getLatest(id primitive.ObjectID) (*model.KVHistory, error) {
	collection := s.c.Database(DB).Collection(CollectionRevision)
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)

	filter := bson.M{"kvID": id.Hex()}

	cur, err := collection.Find(ctx, filter,
		options.Find().SetSort(map[string]interface{}{
			"revision": -1,
		}), options.Find().SetLimit(1))
	if err != nil {
		return nil, err
	}
	h := &model.KVHistory{}
	var exist bool
	for cur.Next(ctx) {
		if err := cur.Decode(h); err != nil {
			openlogging.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		exist = true
		break
	}
	if !exist {
		return nil, nil
	}
	return h, nil
}
func NewMongoService(opts Options) (Service, error) {
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}
	c, err := getClient(opts)
	if err != nil {
		return nil, err
	}
	m := &MongodbService{
		c:       c,
		timeout: opts.Timeout,
	}
	return m, nil
}
func getClient(opts Options) (*mongo.Client, error) {
	if client == nil {
		var err error
		client, err = mongo.NewClient(options.Client().ApplyURI(opts.URI))
		if err != nil {
			return nil, err
		}
		openlogging.Info("connecting to " + opts.URI)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Connect(ctx)
		if err != nil {
			return nil, err
		}
		openlogging.Info("connected to " + opts.URI)
	}
	return client, nil
}
