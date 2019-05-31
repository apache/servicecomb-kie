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

package dao

import (
	"context"

	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//createKey get latest revision from history
//and increase revision of label
//and insert key
func (s *MongodbService) createKey(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	r, err := s.getLatestLabel(ctx, kv.LabelID)
	if err != nil {
		if err != ErrRevisionNotExist {
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
	collection := s.c.Database(DB).Collection(CollectionKV)
	res, err := collection.InsertOne(ctx, kv)
	if err != nil {
		return nil, err
	}
	objectID, _ := res.InsertedID.(primitive.ObjectID)
	kv.ID = objectID
	revision, err := s.AddHistory(ctx, kv.LabelID, kv.Labels, kv.Domain)
	if err != nil {
		openlogging.Warn(
			fmt.Sprintf("can not updateKey version for [%s] [%s] in [%s]",
				kv.Key, kv.Labels, kv.Domain))
	}
	openlogging.Debug(fmt.Sprintf("create %s with labels %s value [%s]", kv.Key, kv.Labels, kv.Value))
	kv.Revision = revision
	return kv, nil

}

//updateKey get latest revision from history
//and increase revision of label
//and updateKey and them add new revision
func (s *MongodbService) updateKey(ctx context.Context, kv *model.KVDoc) (int, error) {
	collection := s.c.Database(DB).Collection(CollectionKV)
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
		fmt.Sprintf("updateKey %s with labels %s value [%s] %d ",
			kv.Key, kv.Labels, kv.Value, ur.ModifiedCount))
	revision, err := s.AddHistory(ctx, kv.LabelID, kv.Labels, kv.Domain)
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
