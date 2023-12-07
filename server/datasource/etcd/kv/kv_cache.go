package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-chassis/foundation/backoff"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
	goCache "github.com/patrickmn/go-cache"
	"go.etcd.io/etcd/api/v3/mvccpb"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/key"
)

func Init() {
	kvCache = NewKvCache()
	go kvCache.Refresh(context.Background())
}

var kvCache *Cache

const (
	prefixKvs            = "kvs"
	cacheExpirationTime  = 10 * time.Minute
	cacheCleanupInterval = 11 * time.Minute
	etcdWatchTimeout     = 1 * time.Hour
	backOffMinInterval   = 5 * time.Second
)

type Cache struct {
	timeOut    time.Duration
	client     etcdadpt.Client
	revision   int64
	kvIDCache  sync.Map
	kvDocCache *goCache.Cache
}

func NewKvCache() *Cache {
	kvDocCache := goCache.New(cacheExpirationTime, cacheCleanupInterval)
	return &Cache{
		timeOut:    etcdWatchTimeout,
		client:     etcdadpt.Instance(),
		revision:   0,
		kvDocCache: kvDocCache,
	}
}

func Enabled() bool {
	return kvCache != nil
}

type CacheSearchReq struct {
	Domain  string
	Project string
	Opts    *datasource.FindOptions
	Regex   *regexp.Regexp
}

func (kc *Cache) Refresh(ctx context.Context) {
	openlog.Info("start to list and watch")
	retries := 0

	timer := time.NewTimer(backOffMinInterval)
	defer timer.Stop()
	for {
		nextPeriod := backOffMinInterval
		if err := kc.listWatch(ctx); err != nil {
			retries++
			nextPeriod = backoff.GetBackoff().Delay(retries)
		} else {
			retries = 0
		}

		select {
		case <-ctx.Done():
			openlog.Info("stop to list and watch")
			return
		case <-timer.C:
			timer.Reset(nextPeriod)
		}
	}
}

func (kc *Cache) listWatch(ctx context.Context) error {
	rsp, err := kc.list(ctx)
	if err != nil {
		return err
	}
	kc.revision = rsp.Revision

	kc.cachePut(rsp)

	return kc.watch(ctx)
}

func (kc *Cache) watch(ctx context.Context) error {
	timoutCtx, cancel := context.WithTimeout(ctx, kc.timeOut)
	defer cancel()

	rev := kc.revision
	opts := append(
		etcdadpt.WatchPrefixOpOptions(prefixKvs),
		etcdadpt.WithRev(kc.revision+1),
		etcdadpt.WithWatchCallback(kc.watchCallBack),
	)
	err := kc.client.Watch(timoutCtx, opts...)
	if err != nil {
		openlog.Error(fmt.Sprintf("watch prefix %s failed, start rev: %d+1->%d->0, err %v", prefixKvs, rev, kc.revision, err))
		kc.revision = 0
	}
	return err
}

func (kc *Cache) list(ctx context.Context) (*etcdadpt.Response, error) {
	rsp, err := kc.client.Do(ctx, etcdadpt.WatchPrefixOpOptions(prefixKvs)...)
	if err != nil {
		openlog.Error(fmt.Sprintf("list prefix %s failed, current rev: %d, err, %v", prefixKvs, kc.revision, err))
		return nil, err
	}
	return rsp, nil
}

func (kc *Cache) watchCallBack(message string, rsp *etcdadpt.Response) error {
	if rsp == nil || len(rsp.Kvs) == 0 {
		return fmt.Errorf("unknown event")
	}
	kc.revision = rsp.Revision

	switch rsp.Action {
	case etcdadpt.ActionPut:
		kc.cachePut(rsp)
	case etcdadpt.ActionDelete:
		kc.cacheDelete(rsp)
	default:
		openlog.Warn(fmt.Sprintf("unrecognized action::%v", rsp.Action))
	}
	return nil
}

func (kc *Cache) cachePut(rsp *etcdadpt.Response) {
	for _, kv := range rsp.Kvs {
		kvDoc, err := kc.GetKvDoc(kv)
		if err != nil {
			openlog.Error(fmt.Sprintf("failed to unmarshal kv, err %v", err))
			continue
		}
		kc.StoreKvDoc(kvDoc.ID, kvDoc)
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project, kvDoc.Labels)
		m, ok := kc.LoadKvIDSet(cacheKey)
		if !ok {
			z := sync.Map{}
			z.Store(kvDoc.ID, struct{}{})
			kc.StoreKvIDSet(cacheKey, z)
			openlog.Info("cacheKey " + cacheKey + "not exists")
			continue
		}
		m.Store(kvDoc.ID, struct{}{})
	}
}

func (kc *Cache) cacheDelete(rsp *etcdadpt.Response) {
	for _, kv := range rsp.Kvs {
		kvDoc, err := kc.GetKvDoc(kv)
		if err != nil {
			openlog.Error(fmt.Sprintf("failed to unmarshal kv, err %v", err))
			continue
		}
		kc.DeleteKvDoc(kvDoc.ID)
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project, kvDoc.Labels)
		m, ok := kc.LoadKvIDSet(cacheKey)
		if !ok {
			openlog.Error("cacheKey " + cacheKey + "not exists")
			continue
		}
		m.Delete(kvDoc.ID)
	}
}

func (kc *Cache) LoadKvIDSet(cacheKey string) (sync.Map, bool) {
	val, ok := kc.kvIDCache.Load(cacheKey)
	if !ok {
		return sync.Map{}, false
	}
	kvIds, ok := val.(sync.Map)
	if !ok {
		return sync.Map{}, false
	}
	return kvIds, true
}

func (kc *Cache) StoreKvIDSet(cacheKey string, kvIds sync.Map) {
	kc.kvIDCache.Store(cacheKey, kvIds)
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

func (kc *Cache) StoreKvDoc(kvID string, kvDoc *model.KVDoc) {
	kc.kvDocCache.SetDefault(kvID, kvDoc)
}

func (kc *Cache) DeleteKvDoc(kvID string) {
	kc.kvDocCache.Delete(kvID)
}

func Search(ctx context.Context, req *CacheSearchReq) (*model.KVResponse, bool, error) {
	if !req.Opts.ExactLabels {
		return nil, false, nil
	}

	openlog.Debug(fmt.Sprintf("using cache to search kv, domain %v, project %v, opts %+v", req.Domain, req.Project, *req.Opts))
	result := &model.KVResponse{
		Data: []*model.KVDoc{},
	}
	cacheKey := kvCache.GetCacheKey(req.Domain, req.Project, req.Opts.Labels)
	kvIds, ok := kvCache.LoadKvIDSet(cacheKey)
	if !ok {
		kvCache.StoreKvIDSet(cacheKey, sync.Map{})
		return result, true, nil
	}

	var docs []*model.KVDoc

	var kvIdsLeft []string
	kvIds.Range(func(kvID, value any) bool {
		if doc, ok := kvCache.LoadKvDoc(kvID.(string)); ok {
			docs = append(docs, doc)
		} else {
			kvIdsLeft = append(kvIdsLeft, kvID.(string))
		}
		return true
	})
	tpData := kvCache.getKvFromEtcd(ctx, req, kvIdsLeft)
	docs = append(docs, tpData...)

	for _, doc := range docs {
		if isMatch(req, doc) {
			datasource.ClearPart(doc)
			result.Data = append(result.Data, doc)
		}
	}
	result.Total = len(result.Data)
	return result, true, nil
}

func (kc *Cache) getKvFromEtcd(ctx context.Context, req *CacheSearchReq, kvIdsLeft []string) []*model.KVDoc {
	if len(kvIdsLeft) == 0 {
		return nil
	}

	openlog.Debug("get kv from etcd by kvId")
	wg := sync.WaitGroup{}
	docs := make([]*model.KVDoc, len(kvIdsLeft))
	for i, kvID := range kvIdsLeft {
		wg.Add(1)
		go func(kvID string, cnt int) {
			defer wg.Done()

			docKey := key.KV(req.Domain, req.Project, kvID)
			kv, err := etcdadpt.Get(ctx, docKey)
			if err != nil {
				openlog.Error(fmt.Sprintf("failed to get kv from etcd, err %v", err))
				return
			}

			doc, err := kc.GetKvDoc(kv)
			if err != nil {
				openlog.Error(fmt.Sprintf("failed to unmarshal kv, err %v", err))
				return
			}

			kc.StoreKvDoc(doc.ID, doc)
			docs[cnt] = doc
		}(kvID, i)
	}
	wg.Wait()
	return docs
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
	return true
}

func (kc *Cache) GetKvDoc(kv *mvccpb.KeyValue) (*model.KVDoc, error) {
	kvDoc := &model.KVDoc{}
	err := json.Unmarshal(kv.Value, kvDoc)
	if err != nil {
		return nil, err
	}
	return kvDoc, nil
}

func (kc *Cache) GetCacheKey(domain, project string, labels map[string]string) string {
	labelFormat := stringutil.FormatMap(labels)
	inputKey := strings.Join([]string{
		"",
		domain,
		project,
		labelFormat,
	}, "/")
	return inputKey
}
