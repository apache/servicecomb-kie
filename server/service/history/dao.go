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
	"github.com/apache/servicecomb-kie/server/db"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//AddHistory increment labels revision and save current label stats to history, then update current revision to db
func AddHistory(ctx context.Context,
	labelRevision *model.LabelRevisionDoc, labelID string, kvs []*model.KVDoc) (int, error) {
	c, err := db.GetClient()
	if err != nil {
		return 0, err
	}
	labelRevision.Revision = labelRevision.Revision + 1

	//save current kv states
	labelRevision.KVs = kvs
	//clear prev id
	labelRevision.ID = primitive.NilObjectID
	collection := c.Database(db.Name).Collection(db.CollectionLabelRevision)
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
	labelCollection := c.Database(db.Name).Collection(db.CollectionLabel)
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
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionLabelRevision)
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
		return nil, db.ErrRevisionNotExist
	}
	return rvs, nil
}
