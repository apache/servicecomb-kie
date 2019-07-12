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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongodbService) getLatestLabel(ctx context.Context, labelID string) (*model.LabelRevisionDoc, error) {
	collection := s.c.Database(DB).Collection(CollectionLabelRevision)
	ctx, _ = context.WithTimeout(ctx, s.timeout)

	filter := bson.M{"label_id": labelID}

	cur, err := collection.Find(ctx, filter,
		options.Find().SetSort(map[string]interface{}{
			"revision": -1,
		}), options.Find().SetLimit(1))
	if err != nil {
		return nil, err
	}
	h := &model.LabelRevisionDoc{}
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
		return nil, ErrRevisionNotExist
	}
	return h, nil
}

//getAndAddHistory get latest labels revision and call addHistory
func (s *MongodbService) getAndAddHistory(ctx context.Context, labelID string, labels map[string]string, domain string) (int, error) {
	r, err := s.getLatestLabel(ctx, labelID)
	if err != nil {
		if err == ErrRevisionNotExist {
			openlogging.Warn(fmt.Sprintf("label revision not exists, create first label revision"))
			r = &model.LabelRevisionDoc{
				LabelID:  labelID,
				Labels:   labels,
				Domain:   domain,
				Revision: 0,
			}
		} else {
			openlogging.Error(fmt.Sprintf("get latest [%s] in [%s],err: %s",
				labelID, domain, err.Error()))
			return 0, err
		}

	}
	r.Revision, err = s.addHistory(ctx, r, labelID)
	if err != nil {
		return 0, err
	}
	return r.Revision, nil
}

//addHistory labels revision plus 1 and save current label stats to history, then update current revision to db
func (s *MongodbService) addHistory(ctx context.Context, labelRevision *model.LabelRevisionDoc, labelID string) (int, error) {
	labelRevision.Revision = labelRevision.Revision + 1
	kvs, err := s.findKeys(ctx, bson.M{"label_id": labelID}, true)
	//Key may be empty When delete
	if err != nil && err != ErrKeyNotExists {
		return 0, err
	}
	//save current kv states
	labelRevision.KVs = kvs
	//clear prev id
	labelRevision.ID = primitive.NilObjectID
	collection := s.c.Database(DB).Collection(CollectionLabelRevision)
	_, err = collection.InsertOne(ctx, labelRevision)
	if err != nil {
		openlogging.Error(err.Error())
		return 0, err
	}
	hex, err := primitive.ObjectIDFromHex(labelID)
	if err != nil {
		openlogging.Error(fmt.Sprintf("convert %s,err:%s", labelID, err))
		return 0, err
	}
	labelCollection := s.c.Database(DB).Collection(CollectionLabel)
	_, err = labelCollection.UpdateOne(ctx, bson.M{"_id": hex}, bson.D{
		{"$set", bson.D{
			{"revision", labelRevision.Revision},
		}},
	})
	if err != nil {
		return 0, err
	}
	openlogging.Debug(fmt.Sprintf("update revision to %d", labelRevision.Revision))
	return labelRevision.Revision, nil
}
