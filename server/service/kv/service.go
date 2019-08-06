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
	"reflect"
	"time"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/db"
	"github.com/apache/servicecomb-kie/server/service/history"
	"github.com/apache/servicecomb-kie/server/service/label"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//MongodbService operate data in mongodb
type MongodbService struct {
	timeout time.Duration
}

//CreateOrUpdate will create or update a key value record
//it first check label exists or not, and create labels if labels is first posted.
//if label exists, then get its latest revision, and update current revision,
//save the current label and its all key values to history collection
//then check key exists or not, then create or update it
func CreateOrUpdate(ctx context.Context, domain string, kv *model.KVDoc, project string) (*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(ctx, db.Timeout)
	if domain == "" {
		return nil, db.ErrMissingDomain
	}
	if len(kv.Labels) == 0 {
		kv.Labels = map[string]string{
			"default": "default",
		}
	}

	//check whether the projecr has certain labels or not
	labelID, err := label.ProjectHasLabel(ctx, domain, project, kv.Labels)

	var l *model.LabelDoc
	if err != nil {
		if err == db.ErrLabelNotExists {
			l, err = label.CreateLabel(ctx, domain, kv.Labels, project)
			if err != nil {
				openlogging.Error("create label failed", openlogging.WithTags(openlogging.Tags{
					"k":      kv.Key,
					"domain": kv.Domain,
				}))
				return nil, err
			}
			labelID = l.ID
		} else {
			return nil, err
		}

	}
	kv.LabelID = labelID.Hex()
	kv.Project = project
	kv.Domain = domain
	if kv.ValueType == "" {
		kv.ValueType = db.DefaultValueType
	}
	keyID, err := KVExist(ctx, domain, kv.Key, kv.Project, WithLabelID(kv.LabelID))
	if err != nil {
		if err == db.ErrKeyNotExists {
			kv, err := createKey(ctx, kv)
			if err != nil {
				return nil, err
			}
			return kv, nil
		}
		return nil, err
	}
	kv.ID = keyID
	revision, err := updateKeyValue(ctx, kv)
	if err != nil {
		return nil, err
	}
	kv.Revision = revision
	kv.Domain = ""
	return kv, nil

}

//KVExist supports you query by label map or labels id
func KVExist(ctx context.Context, domain, key string, project string, options ...FindOption) (primitive.ObjectID, error) {
	ctx, _ = context.WithTimeout(context.Background(), db.Timeout)
	opts := FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.LabelID != "" {
		kvs, err := FindKVByLabelID(ctx, domain, opts.LabelID, key, project)
		if err != nil {
			return primitive.NilObjectID, err
		}
		return kvs[0].ID, nil
	}
	kvs, err := FindKV(ctx, domain, project, WithExactLabels(), WithLabels(opts.Labels), WithKey(key))
	if err != nil {
		return primitive.NilObjectID, err
	}
	if len(kvs) != 1 {
		return primitive.NilObjectID, db.ErrTooMany
	}

	return kvs[0].Data[0].ID, nil

}

//Delete delete kv,If the labelID is "", query the collection kv to get it
//domain=tenant
//1.delete kv;2.add history
func Delete(kvID string, labelID string, domain string, project string) error {
	ctx, _ := context.WithTimeout(context.Background(), db.Timeout)
	if domain == "" {
		return db.ErrMissingDomain
	}
	if project == "" {
		return db.ErrMissingProject
	}
	hex, err := primitive.ObjectIDFromHex(kvID)
	if err != nil {
		return err
	}
	//if labelID == "",get labelID by kvID
	var kv *model.KVDoc
	if labelID == "" {
		kvArray, err := findOneKey(ctx, bson.M{"_id": hex, "project": project})
		if err != nil {
			return err
		}
		if len(kvArray) > 0 {
			kv = kvArray[0]
			labelID = kv.LabelID
		}
	}
	//get Label and check labelID
	r, err := label.GetLatestLabel(ctx, labelID)
	if err != nil {
		if err == db.ErrRevisionNotExist {
			openlogging.Warn(fmt.Sprintf("failed,kvID and labelID do not match"))
			return db.ErrKvIDAndLabelIDNotMatch
		}
		return err
	}
	//delete kv
	err = DeleteKV(ctx, hex, project)
	if err != nil {
		return err
	}
	kvs, err := findKeys(ctx, bson.M{"label_id": labelID, "project": project}, true)
	//Key may be empty When delete
	if err != nil && err != db.ErrKeyNotExists {
		return err
	}
	//Labels will not be empty when deleted
	revision, err := history.AddHistory(ctx, r, labelID, kvs)
	if err != nil {
		openlogging.Warn("add history failed ,", openlogging.WithTags(openlogging.Tags{
			"kvID":    kvID,
			"labelID": labelID,
			"error":   err.Error(),
		}))
	} else {
		openlogging.Info("add history success,", openlogging.WithTags(openlogging.Tags{
			"kvID":     kvID,
			"labelID":  labelID,
			"revision": revision,
		}))
	}
	return nil
}

//FindKVByLabelID get kvs by key and label id
//key can be empty, then it will return all key values
//if key is given, will return 0-1 key value
func FindKVByLabelID(ctx context.Context, domain, labelID, key string, project string) ([]*model.KVDoc, error) {

	filter := bson.M{"label_id": labelID, "domain": domain, "project": project}
	if key != "" {
		filter["key"] = key
		return findOneKey(ctx, filter)
	}
	return findKeys(ctx, filter, true)

}

//FindKV get kvs by key, labels
//because labels has a a lot of combination,
//you can use WithDepth(0) to return only one kv which's labels exactly match the criteria
func FindKV(ctx context.Context, domain string, project string, options ...FindOption) ([]*model.KVResponse, error) {
	opts := FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.Timeout == 0 {
		opts.Timeout = db.DefaultTimeout
	}
	if domain == "" {
		return nil, db.ErrMissingDomain
	}
	if project == "" {
		return nil, db.ErrMissingProject
	}

	cur, err := findKV(ctx, domain, project, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	kvResp := make([]*model.KVResponse, 0)
	if opts.Depth == 0 && opts.Key != "" {
		openlogging.Debug("find one key", openlogging.WithTags(
			map[string]interface{}{
				"key":    opts.Key,
				"label":  opts.Labels,
				"domain": domain,
			},
		))
		return cursorToOneKV(ctx, cur, opts.Labels)
	}
	openlogging.Debug("find more", openlogging.WithTags(openlogging.Tags{
		"depth":  opts.Depth,
		"k":      opts.Key,
		"labels": opts.Labels,
	}))
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
		openlogging.Debug(fmt.Sprintf("%v", curKV))
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
		return nil, db.ErrKeyNotExists
	}
	return kvResp, nil

}
