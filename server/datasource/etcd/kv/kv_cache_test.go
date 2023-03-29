package kv

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/little-cui/etcdadpt"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

func init() {
	config.Configurations.Cache.Labels = []string{"environment"}
}

func TestCachePut(t *testing.T) {
	type args struct {
		rsp *etcdadpt.Response
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "put 0 kvDoc, cache should store 0 kvDoc",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{}}},
			want: 0,
		},
		{
			name: "put 1 kvDoc, cache should store 1 kvDoc",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
			}}},
			want: 1,
		},
		{
			name: "put 2 kvDocs with different kvIds, cache should store 2 kvDocs",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
				{Value: []byte(`{"id":"2", "key":"withToys", "value":"yes", "labels":{"environment":"testing"}}`)},
			}}},
			want: 2,
		},
		{
			name: "put 2 kvDocs with same kvId, cache should store 1 kvDocs",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
				{Value: []byte(`{"id":"1", "key":"withToys", "value":"yes", "labels":{"environment":"testing"}}`)},
			}}},
			want: 1,
		},
		{
			name: "put 2 kvDoc, but labels are not cached, cache should store 0 kvDoc",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"env":"testing"}}`)},
				{Value: []byte(`{"id":"1", "key":"withToys", "value":"yes", "labels":{"env":"testing"}}`)},
			}}},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := NewKvCache()
			kc.cachePut(tt.args.rsp)
			num := kc.kvDocCache.ItemCount()
			assert.Equal(t, tt.want, num)
		})
	}
}

func TestCacheDelete(t *testing.T) {
	type args struct {
		rsp *etcdadpt.Response
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "first put 2 kvDocs, then delete 0 kvDoc, cache should store 2 kvDocs",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionDelete, Kvs: []*mvccpb.KeyValue{}}},
			want: 2,
		},
		{
			name: "first put 2 kvDocs, then delete kvId=1, cache should store 1 kvDocs",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionDelete, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
			}}},
			want: 1,
		},
		{
			name: "first put 2 kvDocs, then delete kvId=1 and kvId=2, cache should store 0 kvDocs",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
				{Value: []byte(`{"id":"2", "key":"withToys", "value":"yes", "labels":{"environment":"testing"}}`)},
			}}},
			want: 0,
		},
		{
			name: "first put 2 kvDocs, then delete non-exist kvId=0, cache should store 2 kvDocs",
			args: args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"0", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
			}}},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := NewKvCache()
			kc.cachePut(&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
				{Value: []byte(`{"id":"2", "key":"withToys", "value":"yes", "labels":{"environment":"testing"}}`)},
			}})
			kc.cacheDelete(tt.args.rsp)
			num := kc.kvDocCache.ItemCount()
			assert.Equal(t, tt.want, num)
		})
	}
}

func TestWatchCallBack(t *testing.T) {
	type args struct {
		rsp []*etcdadpt.Response
	}
	type want struct {
		kvNum int
		err   error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "receive 2 messages without kvs, expected: error is not nil, cache should store 0 kvDoc",
			args: args{
				rsp: []*etcdadpt.Response{
					{
						Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{},
					},

					{
						Action: etcdadpt.ActionDelete, Kvs: []*mvccpb.KeyValue{},
					},
				},
			},
			want: want{
				kvNum: 0,
				err:   fmt.Errorf("unknown event"),
			},
		},
		{
			name: "receive 1 put message, put 0 kvDoc, expected: error is not nil, cache should store 0 kvDoc",
			args: args{
				rsp: []*etcdadpt.Response{
					{
						Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{},
					},
				},
			},
			want: want{
				kvNum: 0,
				err:   fmt.Errorf("unknown event"),
			},
		},
		{
			name: "receive 1 delete message, delete 0 kvDoc, expected: error is not nil, cache should store 0 kvDoc",
			args: args{
				rsp: []*etcdadpt.Response{{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{}}},
			},
			want: want{
				kvNum: 0,
				err:   fmt.Errorf("unknown event"),
			},
		},
		{
			name: "receive put message, put 1 kvDocs, expected: error is nil, cache should store 1 kvDocs",
			args: args{
				rsp: []*etcdadpt.Response{
					{
						Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)}}},
				},
			},
			want: want{
				kvNum: 1,
				err:   nil,
			},
		},
		{
			name: "receive 1 put message, 1 delete message, first put 1 kvDoc, then delete it, expected: error is nil, cache should store 0 kvDoc",
			args: args{
				rsp: []*etcdadpt.Response{
					{
						Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)}},
					},
					{
						Action: etcdadpt.ActionDelete, Kvs: []*mvccpb.KeyValue{{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)}},
					},
				},
			},
			want: want{
				kvNum: 0,
				err:   nil,
			},
		},
		{
			name: "receive put message put 1 kvDoc, but labels are not cached, cache should store 0 kvDoc",
			args: args{
				[]*etcdadpt.Response{
					{
						Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"env":"testing"}}`)}},
					},
				},
			},
			want: want{
				kvNum: 0,
				err:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := NewKvCache()
			for _, rsp := range tt.args.rsp {
				err := kc.watchCallBack("", rsp)
				assert.Equal(t, tt.want.err, err)
			}
			num := kc.kvDocCache.ItemCount()
			assert.Equal(t, tt.want.kvNum, num)
		})
	}
}

func TestStoreAndLoadKvDoc(t *testing.T) {
	type want struct {
		kvDoc *model.KVDoc
		exist bool
	}
	type args struct {
		kvID               string
		kvDoc              *model.KVDoc
		expireTime         time.Duration
		waitTimeAfterStore time.Duration
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "store 1 kv and the expire time is 1 seconds, then load the kv with no wait time, expect: load kv successfully",
			args: args{
				kvID: "",
				kvDoc: &model.KVDoc{
					ID:    "1",
					Key:   "withFood",
					Value: "yes",
					Labels: map[string]string{
						"environment": "testing",
					},
				},
				expireTime:         1 * time.Second,
				waitTimeAfterStore: 0,
			},
			want: want{
				kvDoc: &model.KVDoc{
					ID:    "1",
					Key:   "withFood",
					Value: "yes",
					Labels: map[string]string{
						"environment": "testing",
					},
				},
				exist: true,
			},
		},
		{
			name: "store 1 kv and the expire time is 1 seconds, after waiting 2 seconds, then load the kv, expect: unable to load the kv",
			args: args{
				kvID: "",
				kvDoc: &model.KVDoc{
					ID:    "1",
					Key:   "withFood",
					Value: "yes",
					Labels: map[string]string{
						"environment": "testing",
					},
				},
				expireTime:         1 * time.Second,
				waitTimeAfterStore: 2 * time.Second,
			},
			want: want{
				kvDoc: nil,
				exist: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := NewKvCache()
			kc.kvDocCache = goCache.New(tt.args.expireTime, tt.args.expireTime)
			kc.StoreKvDoc(tt.args.kvID, tt.args.kvDoc)
			time.Sleep(tt.args.waitTimeAfterStore)
			doc, exist := kc.LoadKvDoc(tt.args.kvID)
			assert.Equal(t, tt.want.exist, exist)
			reflect.DeepEqual(tt.want.kvDoc, doc)
		})
	}
}
