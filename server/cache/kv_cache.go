package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/key"
	"github.com/go-chassis/foundation/backoff"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
	goCache "github.com/patrickmn/go-cache"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

func Init() {
	Kc = NewKvCache()
	go Kc.Refresh(context.Background())
}

var Kc *KvCache

const (
	PrefixKvs            = "kvs"
	cacheExpirationTime  = 10 * time.Minute
	cacheCleanupInterval = 11 * time.Minute
	etcdWatchTimeout     = 1 * time.Hour
	backOffMinInterval   = 5 * time.Second
)

type KvIdSet map[string]struct{}

type KvCache struct {
	timeOut    time.Duration
	client     etcdadpt.Client
	revision   int64
	kvIdCache  sync.Map
	kvDocCache *goCache.Cache
}

type KvCacheSearchReq struct {
	Domain  string
	Project string
	Opts    *datasource.FindOptions
	Regex   *regexp.Regexp
}

func NewKvCache() *KvCache {
	kvDocCache := goCache.New(cacheExpirationTime, cacheCleanupInterval)
	return &KvCache{
		timeOut:    etcdWatchTimeout,
		client:     etcdadpt.Instance(),
		revision:   0,
		kvDocCache: kvDocCache,
	}
}

func Enabled() bool {
	return Kc != nil
}

func (kc *KvCache) Refresh(ctx context.Context) {
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

func (kc *KvCache) listWatch(ctx context.Context) error {
	rsp, err := kc.list(ctx)
	if err != nil {
		return err
	}
	kc.revision = rsp.Revision

	kc.cachePut(rsp)

	return kc.watch(ctx)
}

func (kc *KvCache) watch(ctx context.Context) error {
	timoutCtx, cancel := context.WithTimeout(ctx, kc.timeOut)
	defer cancel()

	rev := kc.revision
	opts := append(
		etcdadpt.WatchPrefixOpOptions(PrefixKvs),
		etcdadpt.WithRev(kc.revision+1),
		etcdadpt.WithWatchCallback(kc.watchCallBack),
	)
	err := kc.client.Watch(timoutCtx, opts...)
	if err != nil {
		openlog.Error(fmt.Sprintf("watch prefix %s failed, start rev: %d+1->%d->0, err %v", PrefixKvs, rev, kc.revision, err))
		kc.revision = 0
	}
	return err
}

func (kc *KvCache) list(ctx context.Context) (*etcdadpt.Response, error) {
	rsp, err := kc.client.Do(ctx, etcdadpt.WatchPrefixOpOptions(PrefixKvs)...)
	if err != nil {
		openlog.Error(fmt.Sprintf("list prefix %s failed, current rev: %d, err, %v", PrefixKvs, kc.revision, err))
		return nil, err
	}
	return rsp, nil
}

func (kc *KvCache) watchCallBack(message string, rsp *etcdadpt.Response) error {
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

func (kc *KvCache) cachePut(rsp *etcdadpt.Response) {
	for _, kv := range rsp.Kvs {
		kvDoc, err := kc.GetKvDoc(kv)
		if err != nil {
			openlog.Error(fmt.Sprintf("failed to unmarshal kv, err %v", err))
			continue
		}
		kc.StoreKvDoc(kvDoc.ID, kvDoc)
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project, kvDoc.Labels)
		m, ok := kc.LoadKvIdSet(cacheKey)
		if !ok {
			kc.StoreKvIdSet(cacheKey, KvIdSet{kvDoc.ID: struct{}{}})
			openlog.Info("cacheKey " + cacheKey + "not exists")
			continue
		}
		m[kvDoc.ID] = struct{}{}
	}
}

func (kc *KvCache) cacheDelete(rsp *etcdadpt.Response) {
	for _, kv := range rsp.Kvs {
		kvDoc, err := kc.GetKvDoc(kv)
		if err != nil {
			openlog.Error(fmt.Sprintf("failed to unmarshal kv, err %v", err))
			continue
		}
		kc.DeleteKvDoc(kvDoc.ID)
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project, kvDoc.Labels)
		m, ok := kc.LoadKvIdSet(cacheKey)
		if !ok {
			openlog.Error("cacheKey " + cacheKey + "not exists")
			continue
		}
		delete(m, kvDoc.ID)
	}
}

func (kc *KvCache) LoadKvIdSet(cacheKey string) (KvIdSet, bool) {
	val, ok := kc.kvIdCache.Load(cacheKey)
	if !ok {
		return nil, false
	}
	kvIds, ok := val.(KvIdSet)
	if !ok {
		return nil, false
	}
	return kvIds, true
}

func (kc *KvCache) StoreKvIdSet(cacheKey string, kvIds KvIdSet) {
	kc.kvIdCache.Store(cacheKey, kvIds)
}

func (kc *KvCache) LoadKvDoc(kvId string) (*model.KVDoc, bool) {
	val, ok := kc.kvDocCache.Get(kvId)
	if !ok {
		return nil, false
	}
	doc, ok := val.(*model.KVDoc)
	if !ok {
		return nil, false
	}
	return doc, true
}

func (kc *KvCache) StoreKvDoc(kvId string, kvDoc *model.KVDoc) {
	kc.kvDocCache.SetDefault(kvId, kvDoc)
}

func (kc *KvCache) DeleteKvDoc(kvId string) {
	kc.kvDocCache.Delete(kvId)
}

func Search(ctx context.Context, req *KvCacheSearchReq) (*model.KVResponse, bool, error) {
	if !req.Opts.ExactLabels {
		return nil, false, nil
	}

	openlog.Debug(fmt.Sprintf("using cache to search kv, domain %v, project %v, opts %+v", req.Domain, req.Project, *req.Opts))
	result := &model.KVResponse{}
	cacheKey := Kc.GetCacheKey(req.Domain, req.Project, req.Opts.Labels)
	kvIds, ok := Kc.LoadKvIdSet(cacheKey)
	if !ok {
		Kc.StoreKvIdSet(cacheKey, KvIdSet{})
		return result, true, nil
	}

	var docs []*model.KVDoc

	var kvIdsLeft []string
	for kvId := range kvIds {
		if doc, ok := Kc.LoadKvDoc(kvId); ok {
			datasource.ClearPart(doc)
			docs = append(docs, doc)
			continue
		}
		kvIdsLeft = append(kvIdsLeft, kvId)
	}

	tpData := Kc.getKvFromEtcd(ctx, req, kvIdsLeft)
	docs = append(docs, tpData...)

	for i := range docs {
		if isMatch(req, docs[i]) {
			result.Data = append(result.Data, docs[i])
		}
	}
	result.Total = len(result.Data)
	return result, true, nil
}

func (kc *KvCache) getKvFromEtcd(ctx context.Context, req *KvCacheSearchReq, kvIdsLeft []string) []*model.KVDoc {
	if len(kvIdsLeft) == 0 {
		return nil
	}

	openlog.Debug("get kv from etcd by kvId")
	wg := sync.WaitGroup{}
	docs := make([]*model.KVDoc, len(kvIdsLeft))
	for i, kvId := range kvIdsLeft {
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
			datasource.ClearPart(doc)
			docs[cnt] = doc
		}(kvId, i)
	}
	wg.Wait()
	return docs
}

func isMatch(req *KvCacheSearchReq, doc *model.KVDoc) bool {
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

func (kc *KvCache) GetKvDoc(kv *mvccpb.KeyValue) (*model.KVDoc, error) {
	kvDoc := &model.KVDoc{}
	err := json.Unmarshal(kv.Value, kvDoc)
	if err != nil {
		return nil, err
	}
	return kvDoc, nil
}

func (kc *KvCache) GetCacheKey(domain, project string, labels map[string]string) string {
	labelFormat := stringutil.FormatMap(labels)
	inputKey := strings.Join([]string{
		"",
		domain,
		project,
		labelFormat,
	}, "/")
	return inputKey
}
