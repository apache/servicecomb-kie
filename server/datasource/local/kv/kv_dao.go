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
	"github.com/apache/servicecomb-kie/server/datasource/local/file"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/util"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/auth"
	"github.com/go-chassis/openlog"
)

// Dao operate data in local
type Dao struct {
}

func (s *Dao) Create(ctx context.Context, kv *model.KVDoc, options ...datasource.WriteOption) (*model.KVDoc, error) {
	if err := auth.CheckCreateKV(ctx, kv); err != nil {
		return nil, err
	}

	err := create(kv)
	if err != nil {
		openlog.Error("create error", openlog.WithTags(openlog.Tags{
			"err": err.Error(),
			"kv":  kv,
		}))
		return nil, err
	}

	return kv, nil
}

func create(kv *model.KVDoc) (err error) {
	data, _ := json.Marshal(&kv)
	rollbackOperations := []file.FileDoRecord{}

	defer func() {
		if err != nil {
			file.Rollback(rollbackOperations)
		}
	}()

	err = file.CreateOrUpdateFile(path.Join(file.FileRootPath, kv.Domain, kv.Project, kv.ID, strconv.FormatInt(kv.UpdateRevision, 10)+".json"), data, &rollbackOperations, false)
	if err != nil {
		return err
	}

	err = file.CreateOrUpdateFile(path.Join(file.FileRootPath, kv.Domain, kv.Project, kv.ID, file.NewstKVFile), data, &rollbackOperations, false)
	return err
}

// Update update key value
func (s *Dao) Update(ctx context.Context, kv *model.KVDoc, options ...datasource.WriteOption) error {
	kvpath := path.Join(file.FileRootPath, kv.Domain, kv.Project, kv.ID, file.NewstKVFile)
	kvInfo, err := file.ReadFile(kvpath)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	if kvInfo == nil {
		return datasource.ErrKeyNotExists
	}
	var oldKV model.KVDoc
	err = json.Unmarshal(kvInfo, &oldKV)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}

	if err := auth.CheckUpdateKV(ctx, &oldKV); err != nil {
		return err
	}

	oldKV.LabelFormat = kv.LabelFormat
	oldKV.Value = kv.Value
	oldKV.Status = kv.Status
	oldKV.Checker = kv.Checker
	oldKV.UpdateTime = kv.UpdateTime
	oldKV.UpdateRevision = kv.UpdateRevision

	err = create(kv)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}

	kvCache.CachePut([]*model.KVDoc{kv})
	return nil
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
	kvs, err := s.listNoAuth(ctx, project, domain,
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

func (s *Dao) GetByKey(ctx context.Context, key, project, domain string, options ...datasource.FindOption) ([]*model.KVDoc, error) {
	opts := datasource.FindOptions{Key: key}
	for _, o := range options {
		o(&opts)
	}
	kvs, err := s.listNoAuth(ctx, project, domain,
		datasource.WithExactLabels(),
		datasource.WithLabels(opts.Labels),
		datasource.WithLabelFormat(opts.LabelFormat),
		datasource.WithKey(key),
		datasource.WithCaseSensitive())
	if err != nil {
		openlog.Error("check kv exist: " + err.Error())
		return nil, err
	}
	if IsUniqueFind(opts) && len(kvs.Data) == 0 {
		return nil, datasource.ErrKeyNotExists
	}
	if len(kvs.Data) != 1 {
		return nil, datasource.ErrTooMany
	}
	return kvs.Data, nil
}

// FindOneAndDelete deletes one kv by id and return the deleted kv as these appeared before deletion
// domain=tenant
func (s *Dao) FindOneAndDelete(ctx context.Context, kvID, project, domain string, options ...datasource.WriteOption) (*model.KVDoc, error) {
	kvDoc := model.KVDoc{}
	kvpath := path.Join(file.FileRootPath, domain, project, kvID, file.NewstKVFile)
	kvFolderPath := path.Join(file.FileRootPath, domain, project, kvID)
	kvTmpFolderPath := path.Join(file.FileRootPath, "tmp", domain, project, kvID)

	kvInfo, err := file.ReadFile(kvpath)
	if err != nil {
		return nil, err
	}

	if kvInfo == nil {
		return nil, datasource.ErrKeyNotExists
	}

	err = file.MoveDir(kvFolderPath, kvTmpFolderPath)

	if err != nil {
		openlog.Error("delete Key error: " + err.Error())
		return nil, err
	}

	err = json.Unmarshal(kvInfo, &kvDoc)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		moveDirErr := file.MoveDir(kvTmpFolderPath, kvFolderPath)
		if moveDirErr != nil {
			openlog.Error("rollback error when delete kv: " + err.Error())
		}
		return nil, err
	}
	err = file.CleanDir(kvTmpFolderPath)
	if err != nil {
		openlog.Warn("clean tmp dir error when delete kv: " + err.Error())
	}
	err = file.CleanDir(kvFolderPath)
	if err != nil {
		openlog.Warn("clean dir error when delete kv: " + err.Error())
	}
	// delete Cache
	kvCache.CacheDelete([]*model.KVDoc{&kvDoc})
	return &kvDoc, nil
}

// FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
func (s *Dao) FindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string, options ...datasource.WriteOption) ([]*model.KVDoc, int64, error) {
	var docs []*model.KVDoc
	var removedIds []string
	kvParentPath := path.Join(file.FileRootPath, domain, project)
	kvTmpParentPath := path.Join(file.FileRootPath, "tmp", domain, project)
	var err error

	defer func() {
		if err != nil {
			for _, id := range removedIds {
				err = file.MoveDir(path.Join(kvTmpParentPath, id), path.Join(kvParentPath, id))
				if err != nil {
					openlog.Warn("move tmp dir to real dir error when delete many kvs: " + err.Error())
				}
				err = file.CleanDir(path.Join(kvTmpParentPath, id))
				if err != nil {
					openlog.Warn("clean tmp dir error when delete many kvs: " + err.Error())
				}
			}
		} else {
			for _, id := range removedIds {
				err = file.CleanDir(path.Join(kvTmpParentPath, id))
				if err != nil {
					openlog.Warn("clean tmp dir error when delete many kvs: " + err.Error())
				}
				err = file.CleanDir(path.Join(kvParentPath, id))
				if err != nil {
					openlog.Warn("clean real dir error when delete many kvs: " + err.Error())
				}
			}
		}
	}()

	for _, id := range kvIDs {
		kvPath := path.Join(kvParentPath, id, file.NewstKVFile)
		kvInfo, kvErr := getKVDoc(kvPath)
		err = kvErr
		if err != nil {
			return nil, 0, err
		}
		docs = append(docs, kvInfo)

		err = file.MoveDir(path.Join(kvParentPath, id), path.Join(kvTmpParentPath, id))
		if err != nil {
			return nil, 0, err
		} else {
			removedIds = append(removedIds, id)
		}
	}

	if len(docs) == 0 {
		return nil, 0, datasource.ErrKeyNotExists
	}
	kvCache.CacheDelete(docs)
	return docs, int64(len(docs)), nil
}

// Get get kv by kv id
func (s *Dao) Get(ctx context.Context, req *model.GetKVRequest) (*model.KVDoc, error) {
	kvpath := path.Join(file.FileRootPath, req.Domain, req.Project, req.ID, file.NewstKVFile)
	curKV, err := getKVDoc(kvpath)
	if err != nil {
		return nil, err
	}
	if err := auth.CheckGetKV(ctx, curKV); err != nil {
		return nil, err
	}
	return curKV, nil
}

func getKVDoc(kvpath string) (*model.KVDoc, error) {
	kvInfo, err := file.ReadFile(kvpath)
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	if kvInfo == nil {
		return nil, datasource.ErrKeyNotExists
	}
	curKV := &model.KVDoc{}
	err = json.Unmarshal(kvInfo, curKV)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}
	return curKV, nil
}

func (s *Dao) Total(ctx context.Context, project, domain string) (int64, error) {
	kvParentPath := path.Join(file.FileRootPath, domain, project)
	total, err := file.CountInDomain(kvParentPath)

	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		openlog.Error("find total number: " + err.Error())
		return 0, err
	}
	return int64(total), nil
}

// List get kv list by key and criteria
func (s *Dao) List(ctx context.Context, project, domain string, options ...datasource.FindOption) (*model.KVResponse, error) {
	result, opts, err := s.listData(ctx, project, domain, options...)
	if err != nil {
		return nil, err
	}

	filterKVs, err := auth.FilterKVList(ctx, result.Data)
	if err != nil {
		return nil, err
	}

	result.Data = filterKVs
	result.Total = len(filterKVs)

	return pagingResult(result, opts), nil
}

func (s *Dao) listNoAuth(ctx context.Context, project, domain string, options ...datasource.FindOption) (*model.KVResponse, error) {
	result, opts, err := s.listData(ctx, project, domain, options...)
	if err != nil {
		return nil, err
	}

	return pagingResult(result, opts), nil
}

// List get kv list by key and criteria
func (s *Dao) listData(ctx context.Context, project, domain string, options ...datasource.FindOption) (*model.KVResponse, datasource.FindOptions, error) {
	opts := datasource.NewDefaultFindOpts()
	for _, o := range options {
		o(&opts)
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	regex, err := toRegex(opts)
	if err != nil {
		return nil, opts, err
	}

	resultInCache, useCache, kvIdsInCache := Search(&CacheSearchReq{
		Domain:  domain,
		Project: project,
		Opts:    &opts,
		Regex:   regex,
	})
	if useCache {
		openlog.Info("Use Cache Find Success")
	}

	result, err := matchLabelsSearchLocally(ctx, domain, project, regex, opts, kvIdsInCache)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.KVResponse{
				Data: []*model.KVDoc{},
			}, opts, nil
		}
		openlog.Error("list kv failed: " + err.Error())
		return nil, opts, err
	}

	if resultInCache != nil {
		result.Data = append(result.Data, resultInCache.Data...)
		result.Total += len(resultInCache.Data)
	}
	return result, opts, nil
}

func matchLabelsSearchLocally(ctx context.Context, domain, project string, regex *regexp.Regexp, opts datasource.FindOptions, kvIdsInCache []string) (*model.KVResponse, error) {
	openlog.Debug("using labels to search kv")
	kvParentPath := path.Join(file.FileRootPath, domain, project)
	kvs, err := file.ReadAllKvsFromProjectFolder(kvParentPath)
	if err != nil {
		return nil, err
	}
	result := &model.KVResponse{
		Data: []*model.KVDoc{},
	}
	var docs []*model.KVDoc
	for _, kv := range kvs {
		var doc model.KVDoc
		err := json.Unmarshal(kv, &doc)
		if err != nil {
			openlog.Error("decode to KVList error: " + err.Error())
			continue
		}
		var exist = false
		for _, v := range kvIdsInCache {
			if v == doc.ID {
				exist = true
				break
			}
		}
		if exist {
			continue
		}

		if !filterMatch(&doc, opts, regex) {
			continue
		}
		bytes, _ := json.Marshal(doc)
		var docDeepCopy model.KVDoc
		err = json.Unmarshal(bytes, &docDeepCopy)
		if err != nil {
			openlog.Error("decode to KVList error: " + err.Error())
			continue
		}
		docs = append(docs, &docDeepCopy)
		datasource.ClearPart(&doc)
		result.Data = append(result.Data, &doc)
		result.Total++

		if IsUniqueFind(opts) {
			break
		}
	}
	kvCache.CachePut(docs)

	return result, nil
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
	datasource.ReverseByPriorityAndUpdateRev(result.Data)

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
	if opts.Key != "" && doc.Key != opts.Key {
		return false
	}
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
	if opts.Value != "" && !strings.Contains(doc.Value, opts.Value) {
		return false
	}
	return true
}
