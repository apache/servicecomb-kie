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
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"time"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/util"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-chassis/openlog"
)

//const of kv service
const (
	MsgFindKvFailed      = "find kv failed, deadline exceeded"
	MsgFindOneKey        = "find one key"
	MsgCreateLabelFailed = "create label failed"
	FmtErrFindKvFailed   = "can not find kv in %s"
)

//Service operate data in mongodb
type Service struct {
	timeout time.Duration
}

//Create will create a key value record
func (s *Service) Create(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(ctx, session.Timeout)
	if kv.Labels == nil {
		kv.Labels = map[string]string{}
	}
	//check whether the project has certain labels or not
	kv.LabelFormat = stringutil.FormatMap(kv.Labels)
	if kv.ValueType == "" {
		kv.ValueType = session.DefaultValueType
	}
	_, err := s.Exist(ctx, kv.Domain, kv.Key, kv.Project, service.WithLabelFormat(kv.LabelFormat))
	if err == nil {
		return nil, session.ErrKVAlreadyExists
	}
	if err != service.ErrKeyNotExists {
		openlog.Error(err.Error())
		return nil, err
	}
	kv, err = createKey(ctx, kv)
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	clearPart(kv)
	return kv, nil
}

//CreateList will create key value list record
func (s *Service) CreateList(ctx context.Context, kvsDoc *model.KVListDoc) (*model.KVListResponseDoc, error) {
	ctx, _ = context.WithTimeout(ctx, session.Timeout)
	kvs := kvsDoc.KVListDoc
	for i := 0; i < len(kvs); i++ {
		if kvs[i].Labels == nil {
			kvs[i].Labels = map[string]string{}
		}

		//check whether the project has certain labels or not
		kvs[i].LabelFormat = stringutil.FormatMap(kvs[i].Labels)
		if kvs[i].ValueType == "" {
			kvs[i].ValueType = session.DefaultValueType
		}
		_, err := s.Exist(ctx, kvs[i].Domain, kvs[i].Key, kvs[i].Project, service.WithLabelFormat(kvs[i].LabelFormat))
		if err == nil {
			return nil, session.ErrKVAlreadyExists
		}

		if err != service.ErrKeyNotExists {
			openlog.Error(err.Error())
			return nil, err
		}

		kvs[i], err = createKey(ctx, kvs[i])
		if err != nil {
			openlog.Error(err.Error())
			return nil, err
		}
		clearPart(kvs[i])
	}
	kvListDoc := &model.KVListResponseDoc{
		Total:     int64(len(kvs)),
		KVListDoc: kvs,
	}
	return kvListDoc, nil
}

//Update will update a key value record
func (s *Service) Update(ctx context.Context, kv *model.UpdateKVRequest) (*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(ctx, session.Timeout)
	getRequest := &model.GetKVRequest{
		Domain:  kv.Domain,
		Project: kv.Project,
		ID:      kv.ID,
	}
	oldKV, err := s.Get(ctx, getRequest)
	if err != nil {
		return nil, err
	}
	if kv.Status != "" {
		oldKV.Status = kv.Status
	}
	if kv.Value != "" {
		oldKV.Value = kv.Value
	}
	err = updateKeyValue(ctx, oldKV)
	if err != nil {
		return nil, err
	}
	clearPart(oldKV)
	return oldKV, nil

}

//UpdateList will update key value list record
func (s *Service) UpdateList(ctx context.Context, kvsDoc *model.UpdateKVListRequest) (*model.KVListResponseDoc, error) {
	ctx, _ = context.WithTimeout(ctx, session.Timeout)
	kvList := kvsDoc.UpdateKVList
	updateKvs := &model.KVListResponseDoc{
		KVListDoc: make([]*model.KVDoc, 0, len(kvList)),
	}
	for i := 0; i < len(kvList); i++ {
		getRequest := &model.GetKVRequest{
			Domain:  kvList[i].Domain,
			Project: kvList[i].Project,
			ID:      kvList[i].ID,
		}

		oldKV, err := s.Get(ctx, getRequest)
		if err != nil {
			return nil, err
		}

		if kvList[i].Status != "" {
			oldKV.Status = kvList[i].Status
		}
		if kvList[i].Value != "" {
			oldKV.Value = kvList[i].Value
		}
		err = updateKeyValue(ctx, oldKV)
		if err != nil {
			return nil, err
		}
		clearPart(oldKV)
		updateKvs.KVListDoc = append(updateKvs.KVListDoc, oldKV)
	}
	updateKvs.Total = int64(len(kvList))
	return updateKvs, nil
}

//Exist supports you query a key value by label map or labels id
func (s *Service) Exist(ctx context.Context, domain, key string, project string, options ...service.FindOption) (*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(context.Background(), session.Timeout)
	opts := service.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.LabelFormat != "" {
		kvs, err := findKVByLabel(ctx, domain, opts.LabelFormat, key, project)
		if err != nil {
			if err != service.ErrKeyNotExists {
				openlog.Error(err.Error())
			}
			return nil, err
		}
		return kvs[0], nil
	}
	kvs, err := s.List(ctx, domain, project,
		service.WithExactLabels(),
		service.WithLabels(opts.Labels),
		service.WithKey(key))
	if err != nil {
		openlog.Error("check kv exist: " + err.Error())
		return nil, err
	}
	if len(kvs.Data) != 1 {
		return nil, session.ErrTooMany
	}

	return kvs.Data[0], nil

}

//FindOneAndDelete deletes one kv by id and return the deleted kv as these appeared before deletion
//domain=tenant
func (s *Service) FindOneAndDelete(ctx context.Context, kvID string, domain string, project string) (*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(context.Background(), session.Timeout)
	return findOneKVAndDelete(ctx, kvID, project, domain)
}

//FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
func (s *Service) FindManyAndDelete(ctx context.Context, kvIDs []string, domain string, project string) ([]*model.KVDoc, error) {
	ctx, _ = context.WithTimeout(context.Background(), session.Timeout)
	return findKVsAndDelete(ctx, kvIDs, project, domain)
}

//List get kv list by key and criteria
func (s *Service) List(ctx context.Context, domain, project string, options ...service.FindOption) (*model.KVResponse, error) {
	opts := service.NewDefaultFindOpts()
	for _, o := range options {
		o(&opts)
	}
	cur, total, err := findKV(ctx, domain, project, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	result := &model.KVResponse{
		Data: []*model.KVDoc{},
	}
	for cur.Next(ctx) {
		curKV := &model.KVDoc{}
		if err := cur.Decode(curKV); err != nil {
			openlog.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		if opts.ExactLabels {
			if !util.IsEquivalentLabel(opts.Labels, curKV.Labels) {
				continue
			}
		}
		clearPart(curKV)
		result.Data = append(result.Data, curKV)
	}
	result.Total = total
	return result, nil
}

//Get get kvs by id
func (s *Service) Get(ctx context.Context, request *model.GetKVRequest) (*model.KVDoc, error) {
	return findKVDocByID(ctx, request.Domain, request.Project, request.ID)
}

//Total return kv record number
func (s *Service) Total(ctx context.Context, domain string) (int64, error) {
	ctx, _ = context.WithTimeout(ctx, session.Timeout)
	return total(ctx, domain)
}
