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
	"fmt"

	"github.com/apache/servicecomb-kie/server/datasource"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/pubsub"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/foundation/validator"
	"github.com/go-chassis/go-chassis/v2/pkg/backends/quota"
	"github.com/go-chassis/openlog"
)

func ListKV(ctx context.Context, request *model.ListKVRequest) (int64, *model.KVResponse, *errsvc.Error) {
	opts := []datasource.FindOption{
		datasource.WithKey(request.Key),
		datasource.WithLabels(request.Labels),
		datasource.WithOffset(request.Offset),
		datasource.WithLimit(request.Limit),
	}
	m := request.Match
	if m == common.PatternExact {
		opts = append(opts, datasource.WithExactLabels())
	}
	if request.Status != "" {
		opts = append(opts, datasource.WithStatus(request.Status))
	}
	rev, err := datasource.GetBroker().GetRevisionDao().GetRevision(ctx, request.Domain)
	if err != nil {
		return rev, nil, config.NewError(config.ErrInternal, err.Error())
	}
	kv, err := datasource.GetBroker().GetKVDao().List(ctx, request.Domain, request.Project, opts...)
	if err != nil {
		openlog.Error("common: " + err.Error())
		return rev, nil, config.NewError(config.ErrInternal, common.MsgDBError)
	}
	return rev, kv, nil
}

func Post(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, *errsvc.Error) {
	if kv.Status == "" {
		kv.Status = common.StatusDisabled
	}
	err := validator.Validate(kv)
	if err != nil {
		return nil, config.NewError(config.ErrInvalidParams, err.Error())
	}
	err = quota.PreCreate("", kv.Domain, "", 1)
	if err != nil {
		if err == quota.ErrReached {
			openlog.Error(fmt.Sprintf("can not create kv %s@%s, due to quota violation", kv.Key, kv.Project))
			return nil, config.NewError(config.ErrNotEnoughQuota, err.Error())
		}
		openlog.Error(err.Error())
		return nil, config.NewError(config.ErrInternal, "quota check failed")
	}
	kv, err = datasource.GetBroker().GetKVDao().Create(ctx, kv)
	if err != nil {
		openlog.Error(fmt.Sprintf("post err:%s", err.Error()))
		if err == datasource.ErrKVAlreadyExists {
			return nil, config.NewError(config.ErrRecordAlreadyExists, err.Error())
		}
		return nil, config.NewError(config.ErrInternal, "create kv failed")
	}
	return kv, nil
}

func Upload(ctx context.Context, request *model.UploadKVRequest) *model.DocRespOfUpload {
	override := request.Override
	kvs := request.KVs
	result := &model.DocRespOfUpload{
		Success: []*model.KVDoc{},
		Failure: []*model.DocFailedOfUpload{},
	}
	strategy := SelectStrategy(override)
	for i, kv := range kvs {
		if kv == nil {
			continue
		}
		kv.Domain = request.Domain
		kv.Project = request.Project
		kv, err := strategy.Execute(ctx, kv)
		if err != nil {
			if err.Code == config.ErrStopUpload {
				appendAbortFailedKVResult(kvs[i:], result)
				break
			}
			appendFailedKVResult(err, kv, result)
			continue
		}

		Publish(kv)
		result.Success = append(result.Success, kv)
	}
	return result
}

func appendFailedKVResult(err *errsvc.Error, kv *model.KVDoc, result *model.DocRespOfUpload) {
	failedKv := &model.DocFailedOfUpload{
		Key:     kv.Key,
		Labels:  kv.Labels,
		ErrCode: err.Code,
		ErrMsg:  err.Detail,
	}
	result.Failure = append(result.Failure, failedKv)
}

func appendAbortFailedKVResult(kvs []*model.KVDoc, result *model.DocRespOfUpload) {
	for _, kv := range kvs {
		failedKv := &model.DocFailedOfUpload{
			Key:     kv.Key,
			Labels:  kv.Labels,
			ErrCode: config.ErrStopUpload,
			ErrMsg:  "stop overriding kvs after reaching the duplicate kv",
		}
		result.Failure = append(result.Failure, failedKv)
	}
}

func Publish(kv *model.KVDoc) {
	err := pubsub.Publish(&pubsub.KVChangeEvent{
		Key:      kv.Key,
		Labels:   kv.Labels,
		Project:  kv.Project,
		DomainID: kv.Domain,
		Action:   pubsub.ActionPut,
	})
	if err != nil {
		openlog.Warn("lost kv change event when post:" + err.Error())
	}
	openlog.Info(fmt.Sprintf("post [%s] success", kv.ID))
}
