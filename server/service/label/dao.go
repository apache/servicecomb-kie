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

package label

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

//CreateLabel create a new label
func CreateLabel(ctx context.Context, domain string, labels map[string]string, project string) (*model.LabelDoc, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	l := &model.LabelDoc{
		Domain:  domain,
		Labels:  labels,
		Project: project,
	}
	collection := c.Database(db.Name).Collection(db.CollectionLabel)
	res, err := collection.InsertOne(ctx, l)
	if err != nil {
		return nil, err
	}
	objectID, _ := res.InsertedID.(primitive.ObjectID)
	l.ID = objectID
	return l, nil
}

//FindLabels find label doc by labels
//if map is empty. will return default labels doc which has no labels
func FindLabels(ctx context.Context, domain string, labels map[string]string) (*model.LabelDoc, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionLabel)

	filter := bson.M{"domain": domain}
	for k, v := range labels {
		filter["labels."+k] = v
	}
	if len(labels) == 0 {
		filter["labels"] = "default" //allow key without labels
	}
	cur, err := collection.Find(ctx, filter)
	if err != nil {

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
	return nil, db.ErrLabelNotExists
}

//GetLatestLabel query revision table and find maximum revision number
func GetLatestLabel(ctx context.Context, labelID string) (*model.LabelRevisionDoc, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionLabelRevision)

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
		return nil, db.ErrRevisionNotExist
	}
	return h, nil
}

//projectHasLabels check whether the project has certain labels
func projectHasLabels(ctx context.Context, domain string, project string, labels map[string]string) (*model.LabelDoc, error) {
	c, err := db.GetClient()
	if err != nil {
		return nil, err
	}
	collection := c.Database(db.Name).Collection(db.CollectionLabel)

	filter := bson.M{"domain": domain, "project": project}
	for k, v := range labels {
		filter["labels."+k] = v
	}
	if len(labels) == 0 {
		filter["labels"] = "default" //allow key without labels
	}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	openlogging.Debug(fmt.Sprintf("find lables [%s] in [%s], project [%s]", labels, domain, project))
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
	return nil, db.ErrLabelNotExists
}
