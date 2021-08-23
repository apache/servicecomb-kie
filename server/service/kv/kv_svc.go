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
	"time"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/concurrency"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/pubsub"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/foundation/validator"
	"github.com/go-chassis/go-chassis/v2/pkg/backends/quota"
	"github.com/go-chassis/openlog"
	"github.com/satori/go.uuid"
)

var sema = concurrency.NewSemaphore(concurrency.DefaultConcurrency)

func ListKV(ctx context.Context, request *model.ListKVRequest) (int64, *model.KVResponse, *errsvc.Error) {
	sema.Acquire()
	defer sema.Release()
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
	kv, err := datasource.GetBroker().GetKVDao().List(ctx, request.Project, request.Domain, opts...)
	if err != nil {
		openlog.Error("common: " + err.Error())
		return rev, nil, config.NewError(config.ErrInternal, common.MsgDBError)
	}
	return rev, kv, nil
}

//Create get latest revision from history
//and increase revision of label
//and insert key
func Create(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, *errsvc.Error) {
	if kv.Status == "" {
		kv.Status = common.StatusDisabled
	}
	err := validator.Validate(kv)
	if err != nil {
		return nil, config.NewError(config.ErrInvalidParams, err.Error())
	}
	err = quota.PreCreate("", kv.Domain, kv.Project, "", 1)
	if err != nil {
		if err == quota.ErrReached {
			openlog.Error(fmt.Sprintf("can not create kv %s@%s, due to quota violation", kv.Key, kv.Project))
			return nil, config.NewError(config.ErrNotEnoughQuota, err.Error())
		}
		openlog.Error(err.Error())
		return nil, config.NewError(config.ErrInternal, "quota check failed")
	}

	if kv.Labels == nil {
		kv.Labels = map[string]string{}
	}
	kv.LabelFormat = stringutil.FormatMap(kv.Labels)
	if kv.ValueType == "" {
		kv.ValueType = datasource.DefaultValueType
	}
	//check whether the project has certain labels or not
	exist, err := datasource.GetBroker().GetKVDao().Exist(ctx, kv.Key, kv.Project, kv.Domain, datasource.WithLabelFormat(kv.LabelFormat))
	if err != nil {
		openlog.Error(err.Error())
		return nil, config.NewError(config.ErrInternal, "create kv failed")
	}
	if exist {
		return kv, config.NewError(config.ErrRecordAlreadyExists, datasource.ErrKVAlreadyExists.Error())
	}
	revision, err := datasource.GetBroker().GetRevisionDao().ApplyRevision(ctx, kv.Domain)
	if err != nil {
		openlog.Error(err.Error())
		return nil, config.NewError(config.ErrInternal, "create kv failed")
	}
	completeKV(kv, revision)
	kv, err = datasource.GetBroker().GetKVDao().Create(ctx, kv)
	if err != nil {
		openlog.Error(fmt.Sprintf("post err:%s", err.Error()))
		return nil, config.NewError(config.ErrInternal, "create kv failed")
	}
	err = datasource.GetBroker().GetHistoryDao().AddHistory(ctx, kv)
	if err != nil {
		openlog.Warn(
			fmt.Sprintf("can not updateKeyValue version for [%s] [%s] in [%s]",
				kv.Key, kv.Labels, kv.Domain))
	}
	openlog.Debug(fmt.Sprintf("create %s with labels %s value [%s]", kv.Key, kv.Labels, kv.Value))
	datasource.ClearPart(kv)
	return kv, nil
}

func completeKV(kv *model.KVDoc, revision int64) {
	kv.ID = uuid.NewV4().String()
	kv.UpdateRevision = revision
	kv.CreateRevision = revision
	now := time.Now().Unix()
	kv.CreateTime = now
	kv.UpdateTime = now
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

//Update update key value and add new revision
func Update(ctx context.Context, kv *model.UpdateKVRequest) (*model.KVDoc, error) {
	oldKV, err := datasource.GetBroker().GetKVDao().Get(ctx, &model.GetKVRequest{
		Domain:  kv.Domain,
		Project: kv.Project,
		ID:      kv.ID,
	})
	if err != nil {
		return nil, err
	}
	if kv.Status != "" {
		oldKV.Status = kv.Status
	}
	if kv.Value != "" {
		oldKV.Value = kv.Value
	}
	oldKV.UpdateTime = time.Now().Unix()
	oldKV.UpdateRevision, err = datasource.GetBroker().GetRevisionDao().ApplyRevision(ctx, kv.Domain)
	if err != nil {
		return nil, err
	}
	err = datasource.GetBroker().GetKVDao().Update(ctx, oldKV)
	if err != nil {
		return nil, err
	}
	openlog.Info(
		fmt.Sprintf("update %s with labels %s value [%s]",
			oldKV.Key, oldKV.Labels, kv.Value))
	err = datasource.GetBroker().GetHistoryDao().AddHistory(ctx, oldKV)
	if err != nil {
		openlog.Error(
			fmt.Sprintf("can not add revision for [%s] [%s] in [%s],err: %s",
				oldKV.Key, oldKV.Labels, kv.Domain, err))
	}
	openlog.Debug(
		fmt.Sprintf("add history %s with labels %s value [%s]",
			oldKV.Key, oldKV.Labels, oldKV.Value))
	datasource.ClearPart(oldKV)
	return oldKV, nil

}

func FindOneAndDelete(ctx context.Context, kvID string, project, domain string) (*model.KVDoc, error) {
	kv, err := datasource.GetBroker().GetKVDao().FindOneAndDelete(ctx, kvID, project, domain)
	if err != nil {
		return nil, err
	}
	openlog.Info(fmt.Sprintf("delete success,kvID=%s", kvID))
	if _, err := datasource.GetBroker().GetRevisionDao().ApplyRevision(ctx, domain); err != nil {
		openlog.Error(fmt.Sprintf("the kv [%s] is deleted, but increase revision failed: [%s]", kvID, err))
		return nil, err
	}
	err = datasource.GetBroker().GetHistoryDao().DelayDeletionTime(ctx, []string{kvID}, project, domain)
	if err != nil {
		openlog.Error(fmt.Sprintf("add delete time to [%s] failed : [%s]", kvID, err))
	}
	return kv, nil
}
func FindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string) ([]*model.KVDoc, error) {
	kvs, deleted, err := datasource.GetBroker().GetKVDao().FindManyAndDelete(ctx, kvIDs, project, domain)
	if err != nil {
		return nil, err
	}
	if int64(len(kvs)) != deleted {
		openlog.Warn(fmt.Sprintf("The count of found and the count of deleted are not equal, found %d, deleted %d", len(kvs), deleted))
	} else {
		openlog.Info(fmt.Sprintf("deleted %d kvs, their ids are %v", deleted, kvIDs))
	}
	if _, err := datasource.GetBroker().GetRevisionDao().ApplyRevision(ctx, domain); err != nil {
		openlog.Error(fmt.Sprintf("kvs [%v] are deleted, but increase revision failed: [%v]", kvIDs, err))
		return nil, err
	}
	err = datasource.GetBroker().GetHistoryDao().DelayDeletionTime(ctx, kvIDs, project, domain)
	if err != nil {
		openlog.Error(fmt.Sprintf("add delete time to kvs [%s] failed : [%s]", kvIDs, err))
	}
	return kvs, nil
}
func Get(ctx context.Context, req *model.GetKVRequest) (*model.KVDoc, error) {
	return datasource.GetBroker().GetKVDao().Get(ctx, req)
}
func List(ctx context.Context, project, domain string, options ...datasource.FindOption) (*model.KVResponse, error) {
	return datasource.GetBroker().GetKVDao().List(ctx, project, domain, options...)
}
