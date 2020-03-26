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

package history

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//const of history
const (
	maxHistoryNum = 100
)

func getHistoryByKeyID(ctx context.Context, filter bson.M, offset, limit int64) (*model.KVResponse, error) {
	collection := session.GetDB().Collection(session.CollectionKVRevision)
	opt := options.Find().SetSort(map[string]interface{}{
		"revision": -1,
	})
	if offset != 0 && limit != 0 {
		opt = opt.SetLimit(limit)
		opt = opt.SetSkip(offset)
	}
	curTotal, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	cur, err := collection.Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	kvs := make([]*model.KVDoc, 0)
	var exist bool
	for cur.Next(ctx) {
		var elem model.KVDoc
		err := cur.Decode(&elem)
		if err != nil {
			openlogging.Error("decode error: " + err.Error())
			return nil, err
		}
		exist = true
		elem.Domain = ""
		elem.Project = ""
		kvs = append(kvs, &elem)
	}
	if !exist {
		return nil, service.ErrRevisionNotExist
	}
	result := &model.KVResponse{
		Data:  kvs,
		Total: int(curTotal),
	}
	return result, nil
}

//AddHistory add kv history
func AddHistory(ctx context.Context, kv *model.KVDoc) error {
	ctx, cancel := context.WithTimeout(ctx, session.Timeout)
	defer cancel()
	collection := session.GetDB().Collection(session.CollectionKVRevision)
	_, err := collection.InsertOne(ctx, kv)
	if err != nil {
		openlogging.Error(err.Error())
		return err
	}
	err = historyRotate(ctx, kv.ID, kv.Project, kv.Domain)
	if err != nil {
		openlogging.Error("history rotate err: " + err.Error())
		return err
	}
	return nil
}

//AddDeleteTime add delete time to all revisions of the kv,
//thus these revisions will be automatically deleted by TTL index.
func AddDeleteTime(ctx context.Context, kvID, project, domain string) error {
	collection := session.GetDB().Collection(session.CollectionKVRevision)
	now := time.Now()
	_, err := collection.UpdateMany(ctx, bson.M{"id": kvID, "project": project, "domain": domain}, bson.D{
		{"$set", bson.D{
			{"delete_time", now},
		}},
	})
	if err != nil {
		return err
	}
	openlogging.Debug(fmt.Sprintf("added delete time [%s] to key [%s]", now.String(), kvID))
	return nil
}

//historyRotate delete historical versions for a key that exceeds the limited number
func historyRotate(ctx context.Context, kvID, project, domain string) error {
	ctx, cancel := context.WithTimeout(ctx, session.Timeout)
	defer cancel()
	filter := bson.M{"id": kvID, "domain": domain, "project": project}
	collection := session.GetDB().Collection(session.CollectionKVRevision)
	curTotal, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}
	if curTotal <= maxHistoryNum {
		return nil
	}
	opt := options.Find().SetSort(map[string]interface{}{
		"update_revision": 1,
	})
	opt = opt.SetLimit(curTotal - maxHistoryNum)
	cur, err := collection.Find(ctx, filter, opt)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return err
	}
	for cur.Next(ctx) {
		curKV := &model.KVDoc{}
		if err := cur.Decode(curKV); err != nil {
			openlogging.Error("decode to KVs error: " + err.Error())
			return err
		}
		_, err := collection.DeleteOne(ctx, bson.M{
			"id":              kvID,
			"domain":          domain,
			"project":         project,
			"update_revision": curKV.UpdateRevision,
		})
		if err != nil {
			return err
		}
		openlogging.Debug("delete overflowed revision", openlogging.WithTags(openlogging.Tags{
			"id":       curKV.ID,
			"key":      curKV.Key,
			"revision": curKV.UpdateRevision,
		}))
	}

	return nil
}
