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
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/label"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//AddHistory increment labels revision and save current label stats to history, then update current revision to db
func AddHistory(ctx context.Context,
	labelRevision *model.LabelRevisionDoc, labelID string, kvs []*model.KVDoc) (int, error) {
	labelRevision.Revision = labelRevision.Revision + 1

	//save current kv states
	labelRevision.KVs = kvs
	//clear prev id
	labelRevision.ID = ""
	collection := session.GetDB().Collection(session.CollectionLabelRevision)
	_, err := collection.InsertOne(ctx, labelRevision)
	if err != nil {
		openlogging.Error(err.Error())
		return 0, err
	}
	hex, err := primitive.ObjectIDFromHex(labelID)
	if err != nil {
		openlogging.Error(fmt.Sprintf("convert %s,err:%s", labelID, err))
		return 0, err
	}
	labelCollection := session.GetDB().Collection(session.CollectionLabel)
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

func getHistoryByLabelID(ctx context.Context, filter bson.M) ([]*model.LabelRevisionDoc, error) {
	collection := session.GetDB().Collection(session.CollectionLabelRevision)
	cur, err := collection.Find(ctx, filter, options.Find().SetSort(map[string]interface{}{
		"revisions": -1,
	}))
	if err != nil {
		return nil, err
	}
	rvs := []*model.LabelRevisionDoc{}
	var exist bool
	for cur.Next(ctx) {
		var elem model.LabelRevisionDoc
		err := cur.Decode(&elem)
		if err != nil {
			openlogging.Error("decode to LabelRevisionDoc error: " + err.Error())
			return nil, err
		}
		exist = true
		clearRevisionKV(&elem)
		rvs = append(rvs, &elem)
	}
	if !exist {
		return nil, service.ErrRevisionNotExist
	}
	return rvs, nil
}

//GetAndAddHistory get latest labels revision and call AddHistory
func GetAndAddHistory(ctx context.Context,
	labelID string, labels map[string]string, kvs []*model.KVDoc, domain string, project string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, session.Timeout)
	defer cancel()
	r, err := label.GetLatestLabel(ctx, labelID)
	if err != nil {
		if err == service.ErrRevisionNotExist {
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
	r.Revision, err = AddHistory(ctx, r, labelID, kvs)
	if err != nil {
		return 0, err
	}
	return r.Revision, nil
}
