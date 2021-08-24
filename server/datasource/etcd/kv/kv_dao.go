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
	"regexp"
	"strings"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/util"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/key"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
)

//Dao operate data in mongodb
type Dao struct {
}

func (s *Dao) Create(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	bytes, err := json.Marshal(kv)
	if err != nil {
		openlog.Error("marshal kv ")
		return nil, err
	}
	ok, err := etcdadpt.InsertBytes(ctx, key.KV(kv.Domain, kv.Project, kv.ID), bytes)
	if err != nil {
		openlog.Error("create error", openlog.WithTags(openlog.Tags{
			"err": err.Error(),
			"kv":  kv,
		}))
		return nil, err
	}
	if !ok {
		openlog.Error("create error", openlog.WithTags(openlog.Tags{
			"err": datasource.ErrKVAlreadyExists.Error(),
			"kv":  kv,
		}))
		return nil, datasource.ErrKVAlreadyExists
	}
	return kv, nil
}

//Update update key value
func (s *Dao) Update(ctx context.Context, kv *model.KVDoc) error {
	keyKv := key.KV(kv.Domain, kv.Project, kv.ID)
	resp, err := etcdadpt.Get(ctx, keyKv)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	if resp == nil {
		return datasource.ErrRecordNotExists
	}

	var old model.KVDoc
	err = json.Unmarshal(resp.Value, &old)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	old.LabelFormat = kv.LabelFormat
	old.Value = kv.Value
	old.Status = kv.Status
	old.Checker = kv.Checker
	old.UpdateTime = kv.UpdateTime
	old.UpdateRevision = kv.UpdateRevision

	bytes, err := json.Marshal(old)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	err = etcdadpt.PutBytes(ctx, keyKv, bytes)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	return nil
}

//Extract key values
func getValue(str string) string {
	rex := regexp.MustCompile(`\(([^)]+)\)`)
	res := rex.FindStringSubmatch(str)
	return res[len(res)-1]
}

//Exist supports you query a key value by label map or labels id
func (s *Dao) Exist(ctx context.Context, key, project, domain string, options ...datasource.FindOption) (bool, error) {
	opts := datasource.FindOptions{Key: key}
	for _, o := range options {
		o(&opts)
	}
	kvs, err := s.List(ctx, project, domain,
		datasource.WithExactLabels(),
		datasource.WithLabels(opts.Labels),
		datasource.WithLabelFormat(opts.LabelFormat),
		datasource.WithKey(key))
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

//FindOneAndDelete deletes one kv by id and return the deleted kv as these appeared before deletion
//domain=tenant
func (s *Dao) FindOneAndDelete(ctx context.Context, kvID, project, domain string) (*model.KVDoc, error) {
	resp, err := etcdadpt.ListAndDelete(ctx, key.KV(domain, project, kvID))
	if err != nil {
		openlog.Error("delete Key error: " + err.Error())
		return nil, err
	}
	if resp.Count == 0 {
		return nil, datasource.ErrKeyNotExists
	}
	var doc model.KVDoc
	err = json.Unmarshal(resp.Kvs[0].Value, &doc)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}
	return &doc, nil
}

//FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
func (s *Dao) FindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string) ([]*model.KVDoc, int64, error) {
	var opOptions []etcdadpt.OpOptions
	for _, id := range kvIDs {
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
	var docs []*model.KVDoc
	for _, kv := range resp.Kvs {
		var doc model.KVDoc
		err := json.Unmarshal(kv.Value, &doc)
		if err != nil {
			return nil, 0, err
		}
		docs = append(docs, &doc)
	}
	return docs, resp.Count, nil
}

//Get get kv by kv id
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

//List get kv list by key and criteria
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
	return pagingResult(result, opts), nil
}

func IsUniqueFind(opts datasource.FindOptions) bool {
	return opts.LabelFormat != "" && opts.Key != ""
}

func toRegex(opts datasource.FindOptions) (*regexp.Regexp, error) {
	var (
		regex *regexp.Regexp
		value string
	)
	if opts.Key != "" {
		switch {
		case strings.HasPrefix(opts.Key, "beginWith("):
			value = "^" + strings.ReplaceAll(getValue(opts.Key), ".", "\\.") + ".*"
		case strings.HasPrefix(opts.Key, "wildcard("):
			value = strings.ReplaceAll(getValue(opts.Key), ".", "\\.")
			value = strings.ReplaceAll(value, "*", ".*")
		default:
			value = "^" + strings.ReplaceAll(opts.Key, ".", "\\.") + "$"
		}
		var err error
		regex, err = regexp.Compile(value)
		if err != nil {
			openlog.Error("invalid wildcard expr: " + err.Error())
			return nil, err
		}
	}
	return regex, nil
}

func pagingResult(result *model.KVResponse, opts datasource.FindOptions) *model.KVResponse {
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
