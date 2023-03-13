package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/key"
	"github.com/go-chassis/foundation/backoff"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"regexp"
	"strings"
	"sync"
	"time"
)

func Init() {
	Kc = NewKvCache(PrefixKvs, 0, &sync.Map{})
	go Kc.refresh(context.Background())
}

var Kc *KvCache

const (
	PrefixKvs = "kvs"
)

type KvIdSet map[string]struct{}

type KvCache struct {
	Client   etcdadpt.Client
	Prefix   string
	Revision int64
	Cache    *sync.Map
}

type KvCacheSearchReq struct {
	Domain  string
	Project string
	Opts    *datasource.FindOptions
	Regex   *regexp.Regexp
}

func NewKvCache(prefix string, rev int64, cache *sync.Map) *KvCache {
	return &KvCache{
		Client:   etcdadpt.Instance(),
		Prefix:   prefix,
		Revision: rev,
		Cache:    cache,
	}
}

func (kc *KvCache) refresh(ctx context.Context) {
	openlog.Info("start to list and watch")
	retries := 0

	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		nextPeriod := time.Second
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
	rsp, err := kc.Client.Do(ctx, etcdadpt.WatchPrefixOpOptions(kc.Prefix)...)
	if err != nil {
		openlog.Error(fmt.Sprintf("list prefix %s failed, current rev: %d, err, %v", kc.Prefix, kc.Revision, err))
		return err
	}
	kc.Revision = rsp.Revision

	kc.cachePut(rsp)

	rev := kc.Revision
	opts := append(
		etcdadpt.WatchPrefixOpOptions(kc.Prefix),
		etcdadpt.WithRev(kc.Revision+1),
		etcdadpt.WithWatchCallback(kc.watchCallBack),
	)
	err = kc.Client.Watch(ctx, opts...)
	if err != nil {
		openlog.Error(fmt.Sprintf("watch prefix %s failed, start rev: %d+1->%d->0, err %v", kc.Prefix, rev, kc.Revision, err))
		kc.Revision = 0
	}
	return err
}

func (kc *KvCache) watchCallBack(message string, rsp *etcdadpt.Response) error {
	if rsp == nil || len(rsp.Kvs) == 0 {
		return fmt.Errorf("unknown event")
	}
	kc.Revision = rsp.Revision

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
			continue
		}
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project, kvDoc.Labels)
		m, ok := kc.Load(cacheKey)
		if !ok {
			kc.Store(cacheKey, KvIdSet{kvDoc.ID: struct{}{}})
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
			continue
		}
		cacheKey := kc.GetCacheKey(kvDoc.Domain, kvDoc.Project, kvDoc.Labels)
		m, ok := kc.Load(cacheKey)
		if !ok {
			openlog.Error("cacheKey " + cacheKey + "not exists")
			continue
		}
		delete(m, kvDoc.ID)
	}
}

func (kc *KvCache) Load(cacheKey string) (KvIdSet, bool) {
	val, ok := kc.Cache.Load(cacheKey)
	if !ok {
		return nil, false
	}
	kvIds, ok := val.(KvIdSet)
	if !ok {
		return nil, false
	}
	return kvIds, true
}

func (kc *KvCache) Store(cacheKey string, kvIds KvIdSet) {
	kc.Cache.Store(cacheKey, kvIds)
}

func (kc *KvCache) Search(ctx context.Context, req *KvCacheSearchReq) (*model.KVResponse, error) {
	openlog.Debug("using cache to search kv")

	if !req.Opts.ExactLabels {
		openlog.Error("opts.ExactLabels is false, shouldn't use cache to search")
		return nil, fmt.Errorf("opts.ExactLabels is false, shouldn't use cache to search")
	}

	cacheKey := kc.GetCacheKey(req.Domain, req.Project, req.Opts.Labels)
	kvIDs, ok := kc.Load(cacheKey)
	if !ok {
		kc.Store(cacheKey, KvIdSet{})
		openlog.Error("cacheKey " + cacheKey + " not exists")
		return nil, fmt.Errorf("cacheKey " + cacheKey + " not exists")
	}

	result := &model.KVResponse{}
	cnt := 0
	wg := sync.WaitGroup{}
	tpData := make([]*model.KVDoc, len(kvIDs))
	for kvID := range kvIDs {
		wg.Add(1)
		go func(kvID string, cnt int) {
			defer wg.Done()
			kv, err := etcdadpt.Get(ctx, key.KV(req.Domain, req.Project, kvID))
			if err != nil {
				openlog.Error("get kv failed: " + err.Error())
				return
			}
			var doc model.KVDoc
			err = json.Unmarshal(kv.Value, &doc)
			if err != nil {
				openlog.Error("decode to KVList error: " + err.Error())
				return
			}

			if req.Opts.Status != "" && doc.Status != req.Opts.Status {
				return
			}
			if req.Regex != nil && !req.Regex.MatchString(doc.Key) {
				return
			}

			datasource.ClearPart(&doc)
			tpData[cnt] = &doc
		}(kvID, cnt)
		cnt++
	}
	wg.Wait()
	for i := range tpData {
		if tpData[i] == nil {
			continue
		}
		result.Data = append(result.Data, tpData[i])
	}
	result.Total = len(result.Data)
	return result, nil
}

func (kc *KvCache) GetKvDoc(kv *mvccpb.KeyValue) (*model.KVDoc, error) {
	kvDoc := &model.KVDoc{}
	err := json.Unmarshal(kv.Value, kvDoc)
	if err != nil {
		openlog.Error(fmt.Sprintf("failed to unmarshal kv, err %v", err))
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
