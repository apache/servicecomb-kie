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
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/label"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
)

//const
const (
	existKvLimit  = 2
	existKvOffset = 0
)

//Service operate data in mongodb
type Service struct {
	timeout time.Duration
}

//CreateOrUpdate will create or update a key value record
//it first check label exists or not, and create labels if labels is first posted.
//if label exists, then get its latest revision, and update current revision,
//save the current label and its all key values to history collection
//then check key exists or not, then create or update it
func (s *Service) CreateOrUpdate(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(ctx, session.Timeout)
	if kv.Domain == "" {
		return nil, session.ErrMissingDomain
	}
	//check whether the project has certain labels or not
	labelID, err := label.Exist(ctx, kv.Domain, kv.Project, kv.Labels)
	if err != nil {
		if err == session.ErrLabelNotExists {
			var l *model.LabelDoc
			l, err = label.CreateLabel(ctx, kv.Domain, kv.Labels, kv.Project)
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
	kv.LabelID = labelID
	if kv.ValueType == "" {
		kv.ValueType = session.DefaultValueType
	}
	oldKV, err := s.Exist(ctx, kv.Domain, kv.Key, kv.Project, service.WithLabelID(kv.LabelID))
	if err != nil {
		if err != service.ErrKeyNotExists {
			openlogging.Error(err.Error())
			return nil, err
		}
		kv, err := createKey(ctx, kv)
		if err != nil {
			openlogging.Error(err.Error())
			return nil, err
		}
		kv.Domain = ""
		kv.Project = ""
		return kv, nil
	}
	kv.ID = oldKV.ID
	kv.CreateRevision = oldKV.CreateRevision
	err = updateKeyValue(ctx, kv)
	if err != nil {
		return nil, err
	}
	kv.Domain = ""
	kv.Project = ""
	return kv, nil

}

//Exist supports you query a key value by label map or labels id
func (s *Service) Exist(ctx context.Context, domain, key string, project string, options ...service.FindOption) (*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(context.Background(), session.Timeout)
	opts := service.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.LabelID != "" {
		kvs, err := findKVByLabelID(ctx, domain, opts.LabelID, key, project)
		if err != nil {
			openlogging.Error(err.Error())
			return nil, err
		}
		return kvs[0], nil
	}
	kvs, err := s.FindKV(ctx, domain, project,
		service.WithExactLabels(),
		service.WithLabels(opts.Labels),
		service.WithKey(key),
		service.WithLimit(existKvLimit),
		service.WithOffset(existKvOffset))
	if err != nil {
		openlogging.Error(err.Error())
		return nil, err
	}
	if len(kvs) != 1 {
		return nil, session.ErrTooMany
	}

	return kvs[0].Data[0], nil

}

//Delete delete kv,If the labelID is "", query the collection kv to get it
//domain=tenant
func (s *Service) Delete(ctx context.Context, kvID string, domain string, project string) error {
	ctx, _ = context.WithTimeout(context.Background(), session.Timeout)
	if domain == "" {
		return session.ErrMissingDomain
	}
	if project == "" {
		return session.ErrMissingProject
	}
	if kvID == "" {
		return errors.New("key id is empty")
	}
	//delete kv
	err := deleteKV(ctx, kvID, project, domain)
	if err != nil {
		openlogging.Error("can not delete key, err:" + err.Error())
		return errors.New("can not delete key")
	}
	return nil
}

//List get kv list by key and criteria
func (s *Service) List(ctx context.Context, domain, project string, options ...service.FindOption) (*model.KVResponse, error) {
	opts := service.NewDefaultFindOpts()
	for _, o := range options {
		o(&opts)
	}
	cur, err := findKV(ctx, domain, project, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	result := &model.KVResponse{}
	for cur.Next(ctx) {
		curKV := &model.KVDoc{}
		if err := cur.Decode(curKV); err != nil {
			openlogging.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		if opts.ExactLabels {
			if !reflect.DeepEqual(opts.Labels, curKV.Labels) {
				continue
			}
		}
		clearPart(curKV)
		result.Data = append(result.Data, curKV)
	}
	if len(result.Data) == 0 {
		return nil, service.ErrKeyNotExists
	}
	return result, nil
}

//FindKV get kvs by key, labels
//because labels has a a lot of combination,
//you can use WithDepth(0) to return only one kv which's labels exactly match the criteria
func (s *Service) FindKV(ctx context.Context, domain string, project string, options ...service.FindOption) ([]*model.KVResponse, error) {
	opts := service.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.Timeout == 0 {
		opts.Timeout = session.DefaultTimeout
	}
	if domain == "" {
		return nil, session.ErrMissingDomain
	}
	if project == "" {
		return nil, session.ErrMissingProject
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
				clearAll(curKV)
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
			clearAll(curKV)
			openlogging.Debug("add new label group")
			kvResp = append(kvResp, labelGroup)
		}

	}
	if len(kvResp) == 0 {
		return nil, service.ErrKeyNotExists
	}
	return kvResp, nil

}
