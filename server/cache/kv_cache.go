package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/foundation/backoff"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
	"time"
)

func init() {
	go Kc.refresh(context.Background())
}

var Kc = NewKvCache(PrefixKvs, 0, labelsToIDsMap{})

const (
	PrefixKvs = "/kvs"
)

type labelsToIDsMap map[string][]string

type KvCache struct {
	Client   etcdadpt.Client
	Prefix   string
	Revision int64
	Cache    labelsToIDsMap
}

func NewKvCache(prefix string, rev int64, cache labelsToIDsMap) *KvCache {
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

	for _, kv := range rsp.Kvs {
		kvDoc := &model.KVDoc{}
		err = json.Unmarshal(kv.Value, kvDoc)
		if err != nil {
			openlog.Error(fmt.Sprintf("failed to unmarshal kv, err %v", err))
			continue
		}
	}

	rev := kc.Revision
	opts := append(
		etcdadpt.WatchPrefixOpOptions(kc.Prefix),
		etcdadpt.WithRev(kc.Revision+1),
		etcdadpt.WithWatchCallback(func(message string, evt *etcdadpt.Response) error {
			if rsp == nil || len(rsp.Kvs) == 0 {
				return fmt.Errorf("unknown event")
			}
			kc.Revision = rsp.Revision
			return nil
		}),
	)
	err = kc.Client.Watch(ctx, opts...)
	if err != nil {
		openlog.Error(fmt.Sprintf("watch prefix %s failed, start rev: %d+1->%d->0, err %v", kc.Prefix, rev, kc.Revision, err))
		kc.Revision = 0
	}
	return err
}
