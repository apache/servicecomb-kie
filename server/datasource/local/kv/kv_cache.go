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
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/go-chassis/openlog"
	goCache "github.com/patrickmn/go-cache"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"regexp"
	"strings"
	"sync"
	"time"
)

type IDSet map[string]struct{}

func Init() {
	kvCache = NewKvCache()
	go kvCache.Refresh(context.Background())
}

const (
	cacheExpirationTime  = 10 * time.Minute
	cacheCleanupInterval = 11 * time.Minute
	backOffMinInterval   = 5 * time.Second
)

var kvCache *Cache

type CacheSearchReq struct {
	Domain  string
	Project string
	Opts    *datasource.FindOptions
	Regex   *regexp.Regexp
}

type Cache struct {
	revision   int64
	kvIDCache  sync.Map
	kvDocCache *goCache.Cache
}

func NewKvCache() *Cache {
	kvDocCache := goCache.New(cacheExpirationTime, cacheCleanupInterval)
	return &Cache{
		revision:   0,
		kvDocCache: kvDocCache,
	}
}

func (kc *Cache) Refresh(ctx context.Context) {
	openlog.Info("start to list and watch")

	timer := time.NewTimer(backOffMinInterval)
	defer timer.Stop()
	for {
		nextPeriod := backOffMinInterval
		select {
		case <-ctx.Done():
			openlog.Info("stop to list and watch")
			return
		case <-timer.C:
			timer.Reset(nextPeriod)
		}
	}
}

func (kc *Cache) GetKvDoc(kv *mvccpb.KeyValue) (*model.KVDoc, error) {
	kvDoc := &model.KVDoc{}
	err := json.Unmarshal(kv.Value, kvDoc)
	if err != nil {
		return nil, err
	}
	return kvDoc, nil
}

func (kc *Cache) GetCacheKey(domain, project string) string {
	inputKey := strings.Join([]string{
		"",
		domain,
		project,
	}, "/")
	return inputKey
}

func (kc *Cache) StoreKvDoc(kvID string, kvDoc *model.KVDoc) {
	kc.kvDocCache.SetDefault(kvID, kvDoc)
}

func (kc *Cache) StoreKvIDSet(cacheKey string, kvIds IDSet) {
	kc.kvIDCache.Store(cacheKey, kvIds)
}

func (kc *Cache) DeleteKvDoc(kvID string) {
	kc.kvDocCache.Delete(kvID)
}

func (kc *Cache) LoadKvIDSet(cacheKey string) (IDSet, bool) {
	val, ok := kc.kvIDCache.Load(cacheKey)
	if !ok {
		return nil, false
	}
	kvIds, ok := val.(IDSet)
	if !ok {
		return nil, false
	}
	return kvIds, true
}

func (kc *Cache) LoadKvDoc(kvID string) (*model.KVDoc, bool) {
	val, ok := kc.kvDocCache.Get(kvID)
	if !ok {
		return nil, false
	}
	doc, ok := val.(*model.KVDoc)
	if !ok {
		return nil, false
	}
	return doc, true
}

func (kc *Cache) CachePut(kvs []*model.KVDoc) {
	for _, kvDoc := range kvs {
		kc.StoreKvDoc(kvDoc.ID, kvDoc)
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project)
		m, ok := kc.LoadKvIDSet(cacheKey)
		if !ok {
			kc.StoreKvIDSet(cacheKey, IDSet{kvDoc.ID: struct{}{}})
			openlog.Info("cacheKey " + cacheKey + "not exists")
			continue
		}
		m[kvDoc.ID] = struct{}{}
	}
}

func (kc *Cache) CacheDelete(kvs []*model.KVDoc) {
	for _, kvDoc := range kvs {
		kc.DeleteKvDoc(kvDoc.ID)
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project)
		m, ok := kc.LoadKvIDSet(cacheKey)
		if !ok {
			openlog.Error("cacheKey " + cacheKey + "not exists")
			continue
		}
		delete(m, kvDoc.ID)
	}
}

func Search(req *CacheSearchReq) (*model.KVResponse, bool, []string) {
	//if !req.Opts.ExactLabels {
	//	return nil, false, nil
	//}

	openlog.Debug(fmt.Sprintf("using cache to search kv, domain %v, project %v, opts %+v", req.Domain, req.Project, *req.Opts))
	result := &model.KVResponse{
		Data: []*model.KVDoc{},
	}
	cacheKey := kvCache.GetCacheKey(req.Domain, req.Project)
	kvIds, ok := kvCache.LoadKvIDSet(cacheKey)
	if !ok {
		kvCache.StoreKvIDSet(cacheKey, IDSet{})
		return result, true, nil
	}

	var docs []*model.KVDoc

	var kvIdsInCache []string
	for kvID := range kvIds {
		if doc, ok := kvCache.LoadKvDoc(kvID); ok {
			docs = append(docs, doc)
			kvIdsInCache = append(kvIdsInCache, kvID)
			continue
		}
	}

	for _, doc := range docs {
		if isMatch(req, doc) {
			bytes, _ := json.Marshal(doc)
			var docDeepCopy model.KVDoc
			json.Unmarshal(bytes, &docDeepCopy)

			datasource.ClearPart(&docDeepCopy)
			result.Data = append(result.Data, &docDeepCopy)
		}
	}
	result.Total = len(result.Data)
	return result, true, kvIdsInCache
}

func isMatch(req *CacheSearchReq, doc *model.KVDoc) bool {
	if doc == nil {
		return false
	}
	if req.Opts.Status != "" && doc.Status != req.Opts.Status {
		return false
	}
	if req.Regex != nil && !req.Regex.MatchString(doc.Key) {
		return false
	}
	if req.Opts.Value != "" && !strings.Contains(doc.Value, req.Opts.Value) {
		return false
	}
	return true
}
