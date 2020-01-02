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
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/history"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

//createKey get latest revision from history
//and increase revision of label
//and insert key
func createKey(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	kv.ID = uuid.NewV4().String()
	kv.Revision = 1
	_, err := collection.InsertOne(ctx, kv)
	if err != nil {
		return nil, err
	}
	err = history.AddHistory(ctx, kv)
	if err != nil {
		openlogging.Warn(
			fmt.Sprintf("can not updateKeyValue version for [%s] [%s] in [%s]",
				kv.Key, kv.Labels, kv.Domain))
	}
	openlogging.Debug(fmt.Sprintf("create %s with labels %s value [%s]", kv.Key, kv.Labels, kv.Value))

	return kv, nil

}

//updateKeyValue update key value and add new revision
func updateKeyValue(ctx context.Context, kv *model.KVDoc) error {
	collection := session.GetDB().Collection(session.CollectionKV)
	ur, err := collection.UpdateOne(ctx, bson.M{"key": kv.Key, "label_id": kv.LabelID}, bson.D{
		{"$set", bson.D{
			{"value", kv.Value},
			{"checker", kv.Checker},
			{"revision", kv.Revision},
		}},
	})
	if err != nil {
		return err
	}
	openlogging.Debug(
		fmt.Sprintf("updateKeyValue %s with labels %s value [%s] %d ",
			kv.Key, kv.Labels, kv.Value, ur.ModifiedCount))
	err = history.AddHistory(ctx, kv)
	if err != nil {
		openlogging.Error(
			fmt.Sprintf("can not add revision for [%s] [%s] in [%s],err: %s",
				kv.Key, kv.Labels, kv.Domain, err))
	}
	openlogging.Debug(
		fmt.Sprintf("add history %s with labels %s value [%s] %d ",
			kv.Key, kv.Labels, kv.Value, ur.ModifiedCount))
	return nil

}

func findKV(ctx context.Context, domain string, project string, opts service.FindOptions) (*mongo.Cursor, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	ctx, _ = context.WithTimeout(ctx, opts.Timeout)
	filter := bson.M{"domain": domain, "project": project}
	if opts.Key != "" {
		filter["key"] = opts.Key
	}
	if len(opts.Labels) != 0 {
		for k, v := range opts.Labels {
			filter["labels."+k] = v
		}
	}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find kv failed, deadline exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": opts.Timeout,
			}))
			return nil, fmt.Errorf("can not find kv in %s", opts.Timeout)
		}
		return nil, err
	}
	return cur, err
}
func findOneKey(ctx context.Context, filter bson.M) ([]*model.KVDoc, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	sr := collection.FindOne(ctx, filter)
	if sr.Err() != nil {
		return nil, sr.Err()
	}
	curKV := &model.KVDoc{}
	err := sr.Decode(curKV)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, service.ErrKeyNotExists
		}
		openlogging.Error("decode error: " + err.Error())
		return nil, err
	}
	return []*model.KVDoc{curKV}, nil
}

//deleteKV by kvID
func deleteKV(ctx context.Context, kvID, project, domain string) error {
	collection := session.GetDB().Collection(session.CollectionKV)
	dr, err := collection.DeleteOne(ctx, bson.M{"id": kvID, "project": project, "domain": domain})
	//check error and delete number
	if err != nil {
		openlogging.Error(fmt.Sprintf("delete [%s] failed : [%s]", kvID, err))
		return err
	}
	if dr.DeletedCount != 1 {
		openlogging.Warn(fmt.Sprintf("failed, may have been deleted,kvID=%s", kvID))
	} else {
		openlogging.Info(fmt.Sprintf("delete success,kvID=%s", kvID))
	}
	return err
}
func findKeys(ctx context.Context, filter bson.M, withoutLabel bool) ([]*model.KVDoc, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find kvs failed, dead line exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": session.Timeout,
			}))
			return nil, fmt.Errorf("can not find keys due to timout")
		}
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	kvs := make([]*model.KVDoc, 0)
	for cur.Next(ctx) {
		curKV := &model.KVDoc{}
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
		return nil, service.ErrKeyNotExists
	}
	return kvs, nil
}

//findKVByLabelID get kvs by key and label id
//key can be empty, then it will return all key values
//if key is given, will return 0-1 key value
func findKVByLabelID(ctx context.Context, domain, labelID, key string, project string) ([]*model.KVDoc, error) {
	filter := bson.M{"label_id": labelID, "domain": domain, "project": project}
	if key != "" {
		filter["key"] = key
		return findOneKey(ctx, filter)
	}
	return findKeys(ctx, filter, true)

}
