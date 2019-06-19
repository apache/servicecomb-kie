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

//Package dao is the data access layer
package dao

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"time"
)

var client *mongo.Client

//const for dao
const (
	DB                      = "kie"
	CollectionLabel         = "label"
	CollectionKV            = "kv"
	CollectionLabelRevision = "label_revision"
	DefaultTimeout          = 5 * time.Second
	DefaultValueType        = "text"
)

//MongodbService operate data in mongodb
type MongodbService struct {
	c       *mongo.Client
	timeout time.Duration
}

//CreateOrUpdate will create or update a key value record
//it first check label exists or not, and create labels if labels is first posted.
//if label exists, then get its latest revision, and update current revision,
//save the current label and its all key values to history collection
//then check key exists or not, then create or update it
func (s *MongodbService) CreateOrUpdate(ctx context.Context, domain string, kv *model.KVDoc) (*model.KVDoc, error) {
	if domain == "" {
		return nil, ErrMissingDomain
	}
	ctx, _ = context.WithTimeout(ctx, DefaultTimeout)
	//check labels exits or not
	labelID, err := s.LabelsExist(ctx, domain, kv.Labels)
	var l *model.LabelDoc
	if err != nil {
		if err == ErrLabelNotExists {
			l, err = s.createLabel(ctx, domain, kv.Labels)
			if err != nil {
				return nil, err
			}
			labelID = l.ID
		} else {
			return nil, err
		}

	}
	kv.LabelID = labelID.Hex()
	kv.Domain = domain
	if kv.ValueType == "" {
		kv.ValueType = DefaultValueType
	}
	keyID, err := s.KVExist(ctx, domain, kv.Key, WithLabelID(kv.LabelID))
	if err != nil {
		if err == ErrKeyNotExists {
			kv, err := s.createKey(ctx, kv)
			if err != nil {
				return nil, err
			}
			return kv, nil
		}
		return nil, err
	}
	kv.ID = keyID
	revision, err := s.updateKey(ctx, kv)
	if err != nil {
		return nil, err
	}
	kv.Revision = revision
	return kv, nil

}

//FindLabels find label doc by labels
//if map is empty. will return default labels doc which has no labels
func (s *MongodbService) FindLabels(ctx context.Context, domain string, labels map[string]string) (*model.LabelDoc, error) {
	collection := s.c.Database(DB).Collection(CollectionLabel)
	ctx, _ = context.WithTimeout(context.Background(), DefaultTimeout)
	filter := bson.M{"domain": domain}
	for k, v := range labels {
		filter["labels."+k] = v
	}
	if len(labels) == 0 {
		filter["labels"] = "default" //allow key without labels
	}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			return nil, ErrAction("find label", filter, fmt.Errorf("can not reach mongodb in %s", s.timeout))
		}
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	openlogging.Debug(fmt.Sprintf("find lables [%s] in [%s]", labels, domain))
	curLabel := &model.LabelDoc{} //reuse this pointer to reduce GC, only clear label
	//check label length to get the exact match
	for cur.Next(ctx) { //although complexity is O(n), but there won't be so much labels
		curLabel.Labels = nil
		err := cur.Decode(curLabel)
		if err != nil {
			openlogging.Error("decode error: " + err.Error())
			return nil, err
		}
		if len(curLabel.Labels) == len(labels) {
			openlogging.Debug("hit exact labels")
			curLabel.Labels = nil //exact match don't need to return labels
			return curLabel, nil
		}

	}
	return nil, ErrLabelNotExists
}

func (s *MongodbService) findKeys(ctx context.Context, filter bson.M, withoutLabel bool) ([]*model.KVDoc, error) {
	collection := s.c.Database(DB).Collection(CollectionKV)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			return nil, ErrAction("find", filter, fmt.Errorf("can not reach mongodb in %s", s.timeout))
		}
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	kvs := make([]*model.KVDoc, 0)
	curKV := &model.KVDoc{} //reduce GC,but need to clear labels
	for cur.Next(ctx) {
		curKV.Labels = nil
		if err := cur.Decode(curKV); err != nil {
			openlogging.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		if withoutLabel {
			curKV.Labels = nil
		}
		kvs = append(kvs, curKV)

	}
	if len(kvs) == 0 {
		return nil, ErrKeyNotExists
	}
	return kvs, nil
}

//FindKVByLabelID get kvs by key and label id
//key can be empty, then it will return all key values
//if key is given, will return 0-1 key value
func (s *MongodbService) FindKVByLabelID(ctx context.Context, domain, labelID, key string) ([]*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(context.Background(), DefaultTimeout)
	filter := bson.M{"label_id": labelID, "domain": domain}
	if key != "" {
		return s.findOneKey(ctx, filter, key)
	}
	return s.findKeys(ctx, filter, true)

}

//FindKV get kvs by key, labels
//because labels has a a lot of combination,
//you can use WithDepth(0) to return only one kv which's labels exactly match the criteria
func (s *MongodbService) FindKV(ctx context.Context, domain string, options ...FindOption) ([]*model.KVResponse, error) {
	opts := FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if domain == "" {
		return nil, ErrMissingDomain
	}

	cur, err := s.findKV(ctx, domain, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	kvResp := make([]*model.KVResponse, 0)
	if opts.Depth == 0 {
		openlogging.Debug("find one key", openlogging.WithTags(
			map[string]interface{}{
				"key":    opts.Key,
				"label":  opts.Labels,
				"domain": domain,
			},
		))
		return cursorToOneKV(ctx, cur, opts.Labels)
	}
	for cur.Next(ctx) {
		curKV := &model.KVDoc{}

		if err := cur.Decode(curKV); err != nil {
			openlogging.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		if (len(curKV.Labels) - len(opts.Labels)) > opts.Depth {
			//because it is query by labels, so result can not be minus
			//so many labels,then continue
			openlogging.Debug("so deep, skip this key")
			continue
		}
		openlogging.Info(fmt.Sprintf("%v", curKV))
		var groupExist bool
		var labelGroup *model.KVResponse
		for _, labelGroup = range kvResp {
			if reflect.DeepEqual(labelGroup.LabelDoc.Labels, curKV.Labels) {
				groupExist = true
				clearKV(curKV)
				labelGroup.Data = append(labelGroup.Data, curKV)
				break
			}

		}
		if !groupExist {
			labelGroup = &model.KVResponse{
				LabelDoc: &model.LabelDocResponse{
					Labels:  curKV.Labels,
					LabelID: curKV.LabelID,
				},
				Data: []*model.KVDoc{curKV},
			}
			clearKV(curKV)
			openlogging.Debug("add new label group")
			kvResp = append(kvResp, labelGroup)
		}

	}
	if len(kvResp) == 0 {
		return nil, ErrKeyNotExists
	}
	return kvResp, nil

}

//DeleteByID delete a key value by collection ID
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

//Delete remove a list of key values for a tenant
//domain=tenant
func (s *MongodbService) Delete(ids []string, domain string) error {
	if len(ids) == 0 {
		openlogging.Warn("delete error,ids is blank")
		return nil
	}
	if domain == "" {
		return ErrMissingDomain
	}
	collection := s.c.Database(DB).Collection(CollectionKV)
	//transfer id
	var oid []primitive.ObjectID
	for _, v := range ids {
		if v == "" {
			openlogging.Warn("ids contains continuous ','")
			continue
		}
		hex, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			openlogging.Error(fmt.Sprintf("convert %s ,err:%s", v, err))
			return err
		}
		oid = append(oid, hex)
	}
	//use in filter
	filter := &bson.M{"_id": &bson.M{"$in": oid}, "domain": domain}
	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeout)
	dr, err := collection.DeleteMany(ctx, filter)
	//check error and delete number
	if err != nil {
		openlogging.Error(fmt.Sprintf("delete [%v] failed : [%s]", filter, err))
		return err
	}
	if dr.DeletedCount != int64(len(oid)) {
		openlogging.Warn(fmt.Sprintf(" The actual number of deletions[%d] is not equal to the parameters[%d].", dr.DeletedCount, len(oid)))
	} else {
		openlogging.Info(fmt.Sprintf("delete success,count=%d", dr.DeletedCount))
	}
	return nil
}

//NewMongoService create a new mongo db service
func NewMongoService(opts Options) (*MongodbService, error) {
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
