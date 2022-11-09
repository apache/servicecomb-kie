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
	"encoding/json"
	"github.com/apache/servicecomb-kie/server/datasource/auth"
	"regexp"
	"strings"

	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/util"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/key"
)

// Dao operate data in mongodb
type Dao struct {
}

func (s *Dao) Create(ctx context.Context, kv *model.KVDoc, options ...datasource.WriteOption) (*model.KVDoc, error) {
	//rbac
	if err := auth.CheckCreateOneKV(ctx, kv); err != nil {
		return nil, err
	}

	opts := datasource.NewWriteOptions(options...)
	var exist bool
	var err error
	if opts.SyncEnable {
		// if syncEnable is true, will create task in a transaction operation
		exist, err = txnCreate(ctx, kv)
	} else {
		exist, err = create(ctx, kv)
	}
	if err != nil {
		openlog.Error("create error", openlog.WithTags(openlog.Tags{
			"err": err.Error(),
			"kv":  kv,
		}))
		return nil, err
	}
	if !exist {
		openlog.Error("create error", openlog.WithTags(openlog.Tags{
			"err": datasource.ErrKVAlreadyExists.Error(),
			"kv":  kv,
		}))
		return nil, datasource.ErrKVAlreadyExists
	}
	return kv, nil
}

func create(ctx context.Context, kv *model.KVDoc) (bool, error) {
	kvBytes, err := json.Marshal(kv)
	if err != nil {
		openlog.Error("fail to marshal kv " + err.Error())
		return false, err
	}
	return etcdadpt.InsertBytes(ctx, key.KV(kv.Domain, kv.Project, kv.ID), kvBytes)
}

func txnCreate(ctx context.Context, kv *model.KVDoc) (bool, error) {
	kvBytes, err := json.Marshal(kv)
	if err != nil {
		openlog.Error("fail to marshal kv " + err.Error())
		return false, err
	}
	task, err := sync.NewTask(kv.Domain, kv.Project, sync.CreateAction, datasource.ConfigResource, kv)
	if err != nil {
		openlog.Error("fail to create task" + err.Error())
		return false, err
	}
	taskBytes, err := json.Marshal(task)
	if err != nil {
		openlog.Error("fail to marshal task ")
		return false, err
	}
	kvOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.KV(kv.Domain, kv.Project, kv.ID)), etcdadpt.WithValue(kvBytes))
	taskOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.TaskKey(kv.Domain, kv.Project, task.ID, task.Timestamp)), etcdadpt.WithValue(taskBytes))
	resp, err := etcdadpt.TxnWithCmp(ctx, []etcdadpt.OpOptions{kvOpPut, taskOpPut},
		etcdadpt.If(etcdadpt.NotExistKey(string(kvOpPut.Key)), etcdadpt.NotExistKey(string(taskOpPut.Key))), nil)
	if err != nil {
		return false, err
	}
	return resp.Succeeded, nil
}

// Update update key value
func (s *Dao) Update(ctx context.Context, kv *model.KVDoc, options ...datasource.WriteOption) error {
	keyKV := key.KV(kv.Domain, kv.Project, kv.ID)
	resp, err := etcdadpt.Get(ctx, keyKV)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	if resp == nil {
		return datasource.ErrKeyNotExists
	}
	var oldKV model.KVDoc
	err = json.Unmarshal(resp.Value, &oldKV)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}

	//rbac
	if err := auth.CheckUpdateOneKV(ctx, &oldKV); err != nil {
		return err
	}

	oldKV.LabelFormat = kv.LabelFormat
	oldKV.Value = kv.Value
	oldKV.Status = kv.Status
	oldKV.Checker = kv.Checker
	oldKV.UpdateTime = kv.UpdateTime
	oldKV.UpdateRevision = kv.UpdateRevision

	opts := datasource.NewWriteOptions(options...)
	if opts.SyncEnable {
		// if syncEnable is true, will create task in a transaction operation
		err = txnUpdate(ctx, kv)
	} else {
		err = update(ctx, &oldKV, options...)
	}
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	return nil
}

func txnUpdate(ctx context.Context, kv *model.KVDoc) error {
	keyKV := key.KV(kv.Domain, kv.Project, kv.ID)
	kvBytes, err := json.Marshal(kv)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	task, err := sync.NewTask(kv.Domain, kv.Project, sync.UpdateAction, datasource.ConfigResource, kv)
	if err != nil {
		openlog.Error("fail to create task" + err.Error())
		return err
	}
	taskBytes, err := json.Marshal(task)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	kvOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(keyKV), etcdadpt.WithValue(kvBytes))
	taskOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.TaskKey(kv.Domain, kv.Project, task.ID, task.Timestamp)), etcdadpt.WithValue(taskBytes))
	return etcdadpt.Txn(ctx, []etcdadpt.OpOptions{kvOpPut, taskOpPut})
}

func update(ctx context.Context, kv *model.KVDoc, options ...datasource.WriteOption) error {
	keyKV := key.KV(kv.Domain, kv.Project, kv.ID)
	kvBytes, err := json.Marshal(kv)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	return etcdadpt.PutBytes(ctx, keyKV, kvBytes)
}

// Extract key values
func getValue(str string) string {
	rex := regexp.MustCompile(`\(([^)]+)\)`)
	res := rex.FindStringSubmatch(str)
	return res[len(res)-1]
}

// Exist supports you query a key value by label map or labels id
func (s *Dao) Exist(ctx context.Context, key, project, domain string, options ...datasource.FindOption) (bool, error) {
	opts := datasource.FindOptions{Key: key}
	for _, o := range options {
		o(&opts)
	}
	kvs, err := s.List(ctx, project, domain,
		datasource.WithExactLabels(),
		datasource.WithLabels(opts.Labels),
		datasource.WithLabelFormat(opts.LabelFormat),
		datasource.WithKey(key),
		datasource.WithCaseSensitive())
	if err != nil {
		openlog.Error("check kv exist: " + err.Error())
		return false, err
	}
	if IsUniqueFind(opts) && len(kvs.Data) == 0 {
		return false, nil
	}
	if len(kvs.Data) != 1 {
		return false, datasource.ErrTooMany
	}
	return true, nil
}

// FindOneAndDelete deletes one kv by id and return the deleted kv as these appeared before deletion
// domain=tenant
func (s *Dao) FindOneAndDelete(ctx context.Context, kvID, project, domain string, options ...datasource.WriteOption) (*model.KVDoc, error) {
	opts := datasource.NewWriteOptions(options...)
	if opts.SyncEnable {
		// if syncEnable is ture, will delete kv, create task and create tombstone in a transaction operation
		return txnFindOneAndDelete(ctx, kvID, project, domain)
	}
	return findOneAndDelete(ctx, kvID, project, domain)
}

func findOneAndDelete(ctx context.Context, kvID, project, domain string) (*model.KVDoc, error) {
	kvKey := key.KV(domain, project, kvID)
	kvDoc := model.KVDoc{}

	//rbac check
	if _, err := getKVDoc(ctx, domain, project, kvID); err != nil {
		return nil, err
	}

	resp, err := etcdadpt.ListAndDelete(ctx, kvKey)
	if err != nil {
		openlog.Error("delete Key error: " + err.Error())
		return nil, err
	}
	if resp.Count == 0 {
		return nil, datasource.ErrKeyNotExists
	}
	err = json.Unmarshal(resp.Kvs[0].Value, &kvDoc)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}
	return &kvDoc, nil
}

// txnFindOneAndDelete is to start transaction when delete KV, will create task and tombstone in a transaction operation
func txnFindOneAndDelete(ctx context.Context, kvID, project, domain string) (*model.KVDoc, error) {
	kvKey := key.KV(domain, project, kvID)
	kvDoc, err := getKVDoc(ctx, domain, project, kvID)
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	task, err := sync.NewTask(domain, project, sync.DeleteAction, datasource.ConfigResource, kvDoc)
	if err != nil {
		openlog.Error("fail to create task" + err.Error())
		return nil, err
	}
	taskBytes, err := json.Marshal(task)
	if err != nil {
		openlog.Error("fail to marshal task" + err.Error())
		return nil, err
	}
	tombstone := sync.NewTombstone(domain, project, datasource.ConfigResource, datasource.TombstoneID(kvDoc))
	tombstoneBytes, err := json.Marshal(tombstone)
	if err != nil {
		openlog.Error("fail to marshal tombstone" + err.Error())
		return nil, err
	}
	kvOpDel := etcdadpt.OpDel(etcdadpt.WithStrKey(kvKey))
	taskOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.TaskKey(domain, project,
		task.ID, task.Timestamp)), etcdadpt.WithValue(taskBytes))
	tombstoneOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.TombstoneKey(domain, project, tombstone.ResourceType, tombstone.ResourceID)), etcdadpt.WithValue(tombstoneBytes))
	err = etcdadpt.Txn(ctx, []etcdadpt.OpOptions{kvOpDel, taskOpPut, tombstoneOpPut})
	if err != nil {
		openlog.Error("find and delete error", openlog.WithTags(openlog.Tags{
			"err": err.Error(),
		}))
		return nil, err
	}
	return kvDoc, nil
}

// getKVDoc is to get kv for delete
func getKVDoc(ctx context.Context, domain, project, kvID string) (*model.KVDoc, error) {
	resp, err := etcdadpt.Get(ctx, key.KV(domain, project, kvID))
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	if resp == nil {
		return nil, datasource.ErrKeyNotExists
	}
	curKV := &model.KVDoc{}
	err = json.Unmarshal(resp.Value, curKV)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}

	//rbac
	if err := auth.CheckDeleteOneKV(ctx, curKV); err != nil {
		return nil, err
	}

	return curKV, nil
}

// FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
func (s *Dao) FindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string, options ...datasource.WriteOption) ([]*model.KVDoc, int64, error) {
	opts := datasource.NewWriteOptions(options...)
	if opts.SyncEnable {
		// if sync enable is true, will delete kvs, create tasks and tombstones
		return txnFindManyAndDelete(ctx, kvIDs, project, domain)
	}
	return findManyAndDelete(ctx, kvIDs, project, domain)
}

func findManyAndDelete(ctx context.Context, kvIDs []string, project, domain string) ([]*model.KVDoc, int64, error) {
	var docs []*model.KVDoc
	var opOptions []etcdadpt.OpOptions
	for _, id := range kvIDs {
		if _, err := getKVDoc(ctx, domain, project, id); err != nil {
			continue
		}
		opOptions = append(opOptions, etcdadpt.OpDel(etcdadpt.WithStrKey(key.KV(domain, project, id))))
	}
	resp, err := etcdadpt.ListAndDeleteMany(ctx, opOptions...)
	if err != nil {
		openlog.Error("find Keys error: " + err.Error())
		return nil, 0, err
	}
	if resp.Count == 0 {
		return nil, 0, datasource.ErrKeyNotExists
	}
	for _, kv := range resp.Kvs {
		var doc model.KVDoc
		err := json.Unmarshal(kv.Value, &doc)
		if err != nil {
			openlog.Error("fail to unmarshal kv" + err.Error())
			return nil, 0, err
		}
		docs = append(docs, &doc)
	}
	return docs, resp.Count, nil
}

// txnFindManyAndDelete is to start transaction when delete KVs, will create tasks and tombstones in a transaction operation
func txnFindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string) ([]*model.KVDoc, int64, error) {
	var docs []*model.KVDoc
	var opOptions []etcdadpt.OpOptions
	kvTotalNum := len(kvIDs)
	docs = make([]*model.KVDoc, kvTotalNum)
	tasks := make([]*sync.Task, kvTotalNum)
	tombstones := make([]*sync.Tombstone, kvTotalNum)
	successKVNum := 0
	for i := 0; i < kvTotalNum; i++ {
		kvDoc, err := getKVDoc(ctx, domain, project, kvIDs[i])
		// if not find the kv, continue
		if err != nil {
			if err == datasource.ErrKeyNotExists {
				openlog.Error(err.Error())
				continue
			}
			return nil, 0, err
		}
		if kvDoc == nil {
			continue
		}
		task, err := sync.NewTask(domain, project, sync.DeleteAction, datasource.ConfigResource, kvDoc)
		if err != nil {
			openlog.Error("fail to create task")
			return nil, 0, err
		}
		docs[successKVNum] = kvDoc
		tasks[successKVNum] = task
		tombstones[successKVNum] = sync.NewTombstone(domain, project, datasource.ConfigResource,
			datasource.TombstoneID(kvDoc))
		successKVNum++
	}
	if successKVNum == 0 {
		return nil, 0, datasource.ErrKeyNotExists
	}
	if successKVNum != kvTotalNum {
		docs = docs[:successKVNum]
		tasks = tasks[:successKVNum]
		tombstones = tombstones[:successKVNum]
	}
	for _, id := range kvIDs {
		opOptions = append(opOptions, etcdadpt.OpDel(etcdadpt.WithStrKey(key.KV(domain, project, id))))
	}
	for _, task := range tasks {
		taskBytes, err := json.Marshal(task)
		if err != nil {
			openlog.Error("fail to marshal task" + err.Error())
			return nil, 0, err
		}
		opOptions = append(opOptions, etcdadpt.OpPut(etcdadpt.WithStrKey(key.TaskKey(domain, project,
			task.ID, task.Timestamp)), etcdadpt.WithValue(taskBytes)))
	}
	for _, tombstone := range tombstones {
		tombstoneBytes, err := json.Marshal(tombstone)
		if err != nil {
			openlog.Error("fail to marshal tombstone" + err.Error())
			return nil, 0, err
		}
		opOptions = append(opOptions, etcdadpt.OpPut(etcdadpt.WithStrKey(key.TombstoneKey(domain, project,
			tombstone.ResourceType, tombstone.ResourceID)), etcdadpt.WithValue(tombstoneBytes)))
	}
	err := etcdadpt.Txn(ctx, opOptions)
	if err != nil {
		openlog.Error("find many and delete error", openlog.WithTags(openlog.Tags{
			"err": err.Error(),
		}))
		return nil, 0, err
	}
	return docs, int64(successKVNum), nil
}

// Get get kv by kv id
func (s *Dao) Get(ctx context.Context, req *model.GetKVRequest) (*model.KVDoc, error) {
	resp, err := etcdadpt.Get(ctx, key.KV(req.Domain, req.Project, req.ID))
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	if resp == nil {
		return nil, datasource.ErrKeyNotExists
	}
	curKV := &model.KVDoc{}
	err = json.Unmarshal(resp.Value, curKV)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}

	//rbac
	if err := auth.CheckGetOneKV(ctx, curKV); err != nil {
		return nil, err
	}

	return curKV, nil
}

func (s *Dao) Total(ctx context.Context, project, domain string) (int64, error) {
	_, total, err := etcdadpt.List(ctx, key.KVList(domain, project), etcdadpt.WithCountOnly())
	if err != nil {
		openlog.Error("find total number: " + err.Error())
		return 0, err
	}
	return total, err
}

// List get kv list by key and criteria
func (s *Dao) List(ctx context.Context, project, domain string, options ...datasource.FindOption) (*model.KVResponse, error) {
	opts := datasource.NewDefaultFindOpts()
	for _, o := range options {
		o(&opts)
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	regex, err := toRegex(opts)
	if err != nil {
		return nil, err
	}
	// TODO may be OOM
	kvs, _, err := etcdadpt.List(ctx, key.KVList(domain, project))
	if err != nil {
		openlog.Error("list kv failed: " + err.Error())
		return nil, err
	}
	result := &model.KVResponse{
		Data: []*model.KVDoc{},
	}
	for _, kv := range kvs {
		var doc model.KVDoc
		err := json.Unmarshal(kv.Value, &doc)
		if err != nil {
			openlog.Error("decode to KVList error: " + err.Error())
			continue
		}
		if !filterMatch(&doc, opts, regex) {
			continue
		}

		datasource.ClearPart(&doc)
		result.Data = append(result.Data, &doc)
		result.Total++

		if IsUniqueFind(opts) {
			break
		}
	}

	filterKVs, err := auth.FilterKVList(ctx, result.Data)
	if err != nil {
		return nil, err
	}

	result.Data = filterKVs

	return pagingResult(result, opts), nil
}

func IsUniqueFind(opts datasource.FindOptions) bool {
	return opts.LabelFormat != "" && opts.Key != ""
}

func toRegex(opts datasource.FindOptions) (*regexp.Regexp, error) {
	var value string
	if opts.Key == "" {
		return nil, nil
	}
	switch {
	case strings.HasPrefix(opts.Key, "beginWith("):
		value = strings.ReplaceAll(getValue(opts.Key), ".", "\\.") + ".*"
	case strings.HasPrefix(opts.Key, "wildcard("):
		value = strings.ReplaceAll(getValue(opts.Key), ".", "\\.")
		value = strings.ReplaceAll(value, "*", ".*")
	default:
		value = strings.ReplaceAll(opts.Key, ".", "\\.")
	}
	value = "^" + value + "$"
	if !opts.CaseSensitive {
		value = "(?i)" + value
	}
	regex, err := regexp.Compile(value)
	if err != nil {
		openlog.Error("invalid wildcard expr: " + value + ", error: " + err.Error())
		return nil, err
	}
	return regex, nil
}

func pagingResult(result *model.KVResponse, opts datasource.FindOptions) *model.KVResponse {
	datasource.ReverseByUpdateRev(result.Data)

	if opts.Limit == 0 {
		return result
	}
	total := int64(result.Total)
	if opts.Offset >= total {
		result.Data = []*model.KVDoc{}
		return result
	}
	end := opts.Offset + opts.Limit
	if end > total {
		end = total
	}
	result.Data = result.Data[opts.Offset:end]
	return result
}

func filterMatch(doc *model.KVDoc, opts datasource.FindOptions, regex *regexp.Regexp) bool {
	if opts.Status != "" && doc.Status != opts.Status {
		return false
	}
	if regex != nil && !regex.MatchString(doc.Key) {
		return false
	}
	if len(opts.Labels) != 0 {
		if opts.ExactLabels && !util.IsEquivalentLabel(opts.Labels, doc.Labels) {
			return false
		}
		if !opts.ExactLabels && !util.IsContainLabel(doc.Labels, opts.Labels) {
			return false
		}
	}
	if opts.LabelFormat != "" && doc.LabelFormat != opts.LabelFormat {
		return false
	}
	return true
}
