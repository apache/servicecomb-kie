package cache

import (
	"github.com/little-cui/etcdadpt"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"testing"
)

type args struct {
	rsp *etcdadpt.Response
}

func TestCachePut(t *testing.T) {
	tests := []struct {
		name string
		args args
		want int
	}{
		{"put 0 kvDoc, cache should store 0 kvDoc",
			args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{}}},
			0,
		},
		{"put 1 kvDoc, cache should store 1 kvDoc",
			args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
			}}},
			1,
		},
		{"put 2 kvDocs with different kvIds, cache should store 2 kvDocs",
			args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
				{Value: []byte(`{"id":"2", "key":"withToys", "value":"yes", "labels":{"environment":"testing"}}`)},
			}}},
			2,
		},
		{"put 2 kvDocs with same kvId, cache should store 1 kvDocs",
			args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
				{Value: []byte(`{"id":"1", "key":"withToys", "value":"yes", "labels":{"environment":"testing"}}`)},
			}}},
			1,
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
	tests := []struct {
		name string
		args args
		want int
	}{
		{"first put 2 kvDocs, then delete 0 kvDoc, cache should store 2 kvDocs",
			args{&etcdadpt.Response{Action: etcdadpt.ActionDelete, Kvs: []*mvccpb.KeyValue{}}},
			2,
		},
		{"first put 2 kvDocs, then delete kvId=1, cache should store 1 kvDocs",
			args{&etcdadpt.Response{Action: etcdadpt.ActionDelete, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
			}}},
			1,
		},
		{"first put 2 kvDocs, then delete kvId=1 and kvId=2, cache should store 0 kvDocs",
			args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"1", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
				{Value: []byte(`{"id":"2", "key":"withToys", "value":"yes", "labels":{"environment":"testing"}}`)},
			}}},
			0,
		},
		{"first put 2 kvDocs, then delete non-exist kvId=0, cache should store 2 kvDocs",
			args{&etcdadpt.Response{Action: etcdadpt.ActionPut, Kvs: []*mvccpb.KeyValue{
				{Value: []byte(`{"id":"0", "key":"withFruit", "value":"no", "labels":{"environment":"testing"}}`)},
			}}},
			2,
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
