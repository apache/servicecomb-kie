package cache

import (
	"sync"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/cari/pkg/errsvc"
)

var pollingCache = &LongPollingCache{}

//LongPollingCache exchange space for time
type LongPollingCache struct {
	m sync.Map
}
type DBResult struct {
	KVs *model.KVResponse
	Err *errsvc.Error
	Rev int64
}

func CachedKV() *LongPollingCache {
	return pollingCache
}

//Read reads the cached query result
//only need to filter by labels if match pattern is exact
func (c *LongPollingCache) Read(topic string) (int64, *model.KVResponse, *errsvc.Error) {
	value, ok := c.m.Load(topic)
	if !ok {
		return 0, nil, nil
	}
	t := value.(*DBResult)
	if t.Err != nil {
		return 0, nil, t.Err
	}
	return t.Rev, t.KVs, nil

}

func (c *LongPollingCache) Write(topic string, r *DBResult) {
	c.m.Store(topic, r)
}
