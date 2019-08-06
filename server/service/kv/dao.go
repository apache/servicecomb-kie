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

package kvsvc

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/db"
	"github.com/apache/servicecomb-kie/server/service/history"
	"github.com/apache/servicecomb-kie/server/service/label"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//createKey get latest revision from history
//and increase revision of label
//and insert key
func createKey(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	r, err := label.GetLatestLabel(ctx, kv.LabelID)
	if err != nil {
		if err != db.ErrRevisionNotExist {
			openlogging.Error(fmt.Sprintf("get latest [%s][%s] in [%s],err: %s",
				kv.Key, kv.Labels, kv.Domain, err.Error()))
			return nil, err
		}
		//the first time labels is created, at this time, labels has no revision yet
		//after first key created, labels got revision 1
		r = &model.LabelRevisionDoc{Revision: 0}
	}
	if r != nil {
		r.Revision = r.Revision + 1
	}
	collection := c.Database(db.Name).Collection(db.CollectionKV)
	res, err := collection.InsertOne(ctx, kv)
	if err != nil {
		return nil, err
	}
	objectID, _ := res.InsertedID.(primitive.ObjectID)
	kv.ID = objectID
	kvs, err := findKeys(ctx, bson.M{"label_id": kv.LabelID}, true)
	//Key may be empty When delete
	if err != nil && err != db.ErrKeyNotExists {
		return nil, err
	}
	revision, err := history.GetAndAddHistory(ctx, kv.LabelID, kv.Labels, kvs, kv.Domain, kv.Project)
	if err != nil {
		openlogging.Warn(
			fmt.Sprintf("can not updateKeyValue version for [%s] [%s] in [%s]",
				kv.Key, kv.Labels, kv.Domain))
	}
	openlogging.Debug(fmt.Sprintf("create %s with labels %s value [%s]", kv.Key, kv.Labels, kv.Value))
	kv.Revision = revision
	return kv, nil

}

//updateKeyValue get latest revision from history
//and increase revision of label
//and updateKeyValue and them add new revision
func updateKeyValue(ctx context.Context, kv *model.KVDoc) (int, error) {
	c, err := db.GetClient()
	if err != nil {
		return 0, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionKV)
	ur, err := collection.UpdateOne(ctx, bson.M{"key": kv.Key, "label_id": kv.LabelID}, bson.D{
		{"$set", bson.D{
			{"value", kv.Value},
			{"checker", kv.Checker},
		}},
	})
	if err != nil {
		return 0, err
	}
	openlogging.Debug(
		fmt.Sprintf("updateKeyValue %s with labels %s value [%s] %d ",
			kv.Key, kv.Labels, kv.Value, ur.ModifiedCount))
	kvs, err := findKeys(ctx, bson.M{"label_id": kv.LabelID}, true)
	//Key may be empty When delete
	if err != nil && err != db.ErrKeyNotExists {
		return 0, err
	}
	revision, err := history.GetAndAddHistory(ctx, kv.LabelID, kv.Labels, kvs, kv.Domain, kv.Project)
	if err != nil {
		openlogging.Warn(
			fmt.Sprintf("can not label revision for [%s] [%s] in [%s],err: %s",
				kv.Key, kv.Labels, kv.Domain, err))
	}
	openlogging.Debug(
		fmt.Sprintf("add history %s with labels %s value [%s] %d ",
			kv.Key, kv.Labels, kv.Value, ur.ModifiedCount))
	return revision, nil

}

func findKV(ctx context.Context, domain string, project string, opts FindOptions) (*mongo.Cursor, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionKV)
	ctx, _ = context.WithTimeout(ctx, opts.Timeout)
	filter := bson.M{"domain": domain, "project": project}
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
				"timeout": opts.Timeout,
			}))
			return nil, fmt.Errorf("can not find kv in %s", opts.Timeout)
		}
		return nil, err
	}
	return cur, err
}
func findOneKey(ctx context.Context, filter bson.M) ([]*model.KVDoc, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionKV)
	sr := collection.FindOne(ctx, filter)
	if sr.Err() != nil {
		return nil, sr.Err()
	}
	curKV := &model.KVDoc{}
	err = sr.Decode(curKV)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, db.ErrKeyNotExists
		}
		openlogging.Error("decode error: " + err.Error())
		return nil, err
	}
	return []*model.KVDoc{curKV}, nil
}

//DeleteKV by kvID
func DeleteKV(ctx context.Context, hexID primitive.ObjectID, project string) error {
	c, err := db.GetClient()
	if err != nil {
		return err
	}
	collection := c.Database(db.Name).Collection(db.CollectionKV)
	dr, err := collection.DeleteOne(ctx, bson.M{"_id": hexID, "project": project})
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
func findKeys(ctx context.Context, filter bson.M, withoutLabel bool) ([]*model.KVDoc, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionKV)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find kvs failed, dead line exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": db.Timeout,
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
		return nil, db.ErrKeyNotExists
	}
	return kvs, nil
}
