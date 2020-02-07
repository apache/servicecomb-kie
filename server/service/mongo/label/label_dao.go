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
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	defaultLabels = "default"
)

//FindLabels find label doc by labels and project, check if the project has certain labels
//if map is empty. will return default labels doc which has no labels
func FindLabels(ctx context.Context, domain, project string, labels map[string]string) (*model.LabelDoc, error) {
	collection := session.GetDB().Collection(session.CollectionLabel)
	filter := bson.M{"domain": domain, "project": project}
	filter["format"] = stringutil.FormatMap(labels) //allow key without labels
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	openlogging.Debug(fmt.Sprintf("find labels [%s] in [%s]", labels, domain))
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
	return nil, session.ErrLabelNotExists
}

//Exist check whether the project has certain label or not and return label ID
func Exist(ctx context.Context, domain string, project string, labels map[string]string) (string, error) {
	l, err := FindLabels(ctx, domain, project, labels)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find label failed, dead line exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": session.Timeout,
			}))
			return "", fmt.Errorf("operation timout %s", session.Timeout)
		}
		return "", err
	}

	return l.ID, nil

}

//CreateLabel create a new label
func CreateLabel(ctx context.Context, label *model.LabelDoc) (*model.LabelDoc, error) {
	label.ID = uuid.NewV4().String()
	label.Format = stringutil.FormatMap(label.Labels)
	collection := session.GetDB().Collection(session.CollectionLabel)
	_, err := collection.InsertOne(ctx, label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

//UpdateLabel update alias
func UpdateLabel(ctx context.Context, label *model.LabelDoc) (*model.LabelDoc, error) {
	collection := session.GetDB().Collection(session.CollectionLabel)
	queryFilter := bson.M{"id": label.ID}
	if label.Alias == "" {
		return nil, service.ErrAliasNotGiven
	}
	updateFilter := bson.D{primitive.E{Key: "$set", Value: bson.M{"alias": label.Alias}}}
	cur := collection.FindOneAndUpdate(ctx, queryFilter, updateFilter)
	if cur.Err() != nil {
		return nil, cur.Err()
	}
	return label, nil
}
