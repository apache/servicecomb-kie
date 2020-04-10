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

package v1_test

import (
	"bytes"
	"encoding/json"
	common2 "github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/validate"
	"github.com/apache/servicecomb-kie/server/config"
	handler2 "github.com/apache/servicecomb-kie/server/handler"
	"github.com/apache/servicecomb-kie/server/plugin/qms"
	"github.com/apache/servicecomb-kie/server/pubsub"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/pkg/backends/quota"
	"github.com/go-chassis/go-chassis/server/restful/restfultest"
	log "github.com/go-chassis/paas-lager"
	"github.com/go-mesh/openlogging"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-kie/server/plugin/qms"
	_ "github.com/apache/servicecomb-kie/server/service/mongo"
)

func init() {
	if err := validate.Init(); err != nil {
		panic(err)
	}
	log.Init(log.Config{
		Writers:       []string{"stdout"},
		LoggerLevel:   "DEBUG",
		LogFormatText: false,
	})
	logger := log.NewLogger("ut")
	openlogging.SetLogger(logger)
	//for UT
	config.Configurations = &config.Config{
		DB:             config.DB{},
		ListenPeerAddr: "127.0.0.1:4000",
		AdvertiseAddr:  "127.0.0.1:4000",
	}
	config.Configurations.DB.URI = "mongodb://kie:123@127.0.0.1:27017/kie"
	err := service.DBInit()
	if err != nil {
		panic(err)
	}
	pubsub.Init()
	pubsub.Start()

	err = quota.Init(quota.Options{
		Plugin: "build-in",
	})
	if err != nil {
		panic(err)
	}
	err = archaius.Init(archaius.WithENVSource(), archaius.WithMemorySource())
	if err != nil {
		panic(err)
	}
}
func TestKVResource_Post(t *testing.T) {
	t.Run("post kv, label is service", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:    "timeout",
			Value:  "1s",
			Labels: map[string]string{"service": "utService"},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/test/kie/kv", bytes.NewBuffer(j))
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, chain)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.KVDoc{}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.NotEmpty(t, data.ID)
		assert.Equal(t, kv.Value, data.Value)
		assert.Equal(t, kv.Labels, data.Labels)
	})
	t.Run("post a different key, which label is same to timeout", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:    "interval",
			Value:  "1s",
			Labels: map[string]string{"service": "utService"},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/test/kie/kv", bytes.NewBuffer(j))
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, chain)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.KVDoc{}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.NotEmpty(t, data.ID)
		assert.Equal(t, kv.Value, data.Value)
		assert.Equal(t, kv.Labels, data.Labels)
	})
	t.Run("post kv,label is service and version", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:   "timeout",
			Value: "1s",
			Labels: map[string]string{
				"service": "utService",
				"version": "1.0.0"},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/test/kie/kv", bytes.NewBuffer(j))
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, chain)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.KVDoc{}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.NotEmpty(t, data.ID)
	})
}
func TestKVResource_List(t *testing.T) {
	t.Run("list kv by service label, should return 3 kvs", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.Data))

	})
	var rev string
	t.Run("list kv by service label, exact match,should return 2 kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService&match=exact", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result.Data))
		rev = resp.Header().Get(common2.HeaderRevision)
	})
	t.Run("list kv by service label, with current rev param,should return 304 ", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService&"+common2.QueryParamRev+"="+rev, nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotModified, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with old rev param,should return latest revision", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService&"+common2.QueryParamRev+"=1", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with wait and old rev param,should return latest revision,no wait", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService&wait=1s&"+common2.QueryParamRev+"=1", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		start := time.Now()
		c.ServeHTTP(resp, r)
		duration := time.Since(start)
		t.Log(duration)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with wait and current rev param,should wait and return 304 ", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService&wait=1s&"+common2.QueryParamRev+"="+rev, nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		start := time.Now()
		c.ServeHTTP(resp, r)
		duration := time.Since(start)
		t.Log(duration)
		assert.Equal(t, http.StatusNotModified, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with wait param,will exceed 1s and return 304", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService&wait=1s", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		start := time.Now()
		c.ServeHTTP(resp, r)
		duration := time.Since(start)
		t.Log(duration)
	})
	t.Run("list kv by service label offset, should return 1kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=service:utService&offset=1&limit=1", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Data))
	})
	t.Run("list kv by service label, with wait and match param,not exact match and return 304", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?label=match:test&wait=10s&match=exact", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain-match", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			kv := &model.KVDoc{
				Key:    "testKey",
				Value:  "val",
				Labels: map[string]string{"dummy": "test", "match": "test"},
			}
			j, _ := json.Marshal(kv)
			r2, _ := http.NewRequest("POST", "/v1/test/kie/kv", bytes.NewBuffer(j))
			noopH2 := &handler2.NoopAuthHandler{}
			chain2, _ := handler.CreateChain(common.Provider, "testchain-match", noopH2.Name())
			r2.Header.Set("Content-Type", "application/json")
			kvr2 := &v1.KVResource{}
			c2, _ := restfultest.New(kvr2, chain2)
			resp2 := httptest.NewRecorder()
			c2.ServeHTTP(resp2, r2)
			body, _ := ioutil.ReadAll(resp2.Body)
			data := &model.KVDoc{}
			err = json.Unmarshal(body, data)
			assert.NotEmpty(t, data.ID)
			wg.Done()
		}()
		start := time.Now()
		c.ServeHTTP(resp, r)
		wg.Wait()
		duration := time.Since(start)
		body, _ := ioutil.ReadAll(resp.Body)
		data := &model.KVDoc{}
		err = json.Unmarshal(body, data)
		assert.Equal(t, 304, resp.Code)
		t.Log(duration)
	})
	t.Run("get one key by label, exact match,should return 1 kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?key=timeout&label=service:utService&match=exact", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Data))
	})
	t.Run("get one key by service label should return 2 kv,delete one", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv?key=timeout&label=service:utService", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result.Data))

		r2, _ := http.NewRequest("DELETE", "/v1/test/kie/kv/"+result.Data[0].ID, nil)
		c2, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp2 := httptest.NewRecorder()
		c2.ServeHTTP(resp2, r2)
		assert.Equal(t, http.StatusNoContent, resp2.Code)

	})
}
func TestKVResource_PutAndGet(t *testing.T) {
	var id string
	kv := &model.KVDoc{
		Key:    "user",
		Value:  "guest",
		Labels: map[string]string{"service": "utService"},
	}
	t.Run("create a kv, the value of user is guest", func(t *testing.T) {
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/test/kie/kv", bytes.NewBuffer(j))
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, chain)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVDoc{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, kv.Value, result.Value)
		id = result.ID
	})
	t.Run("get one key by kv_id", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv/"+id, nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVDoc{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, kv.Value, result.Value)
	})
	kvUpdate := &model.KVDoc{
		Value: "admin",
	}
	t.Run("update the kv, set the value of user to admin", func(t *testing.T) {
		j, _ := json.Marshal(kvUpdate)
		r, _ := http.NewRequest("PUT", "/v1/test/kie/kv/"+id, bytes.NewBuffer(j))
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, chain)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVDoc{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, kvUpdate.Value, result.Value)
	})
	t.Run("get one key by kv_id again", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv/"+id, nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVDoc{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, kvUpdate.Value, result.Value)
	})

	t.Run("quota test, can not create new ", func(t *testing.T) {

		archaius.Set(qms.QuotaConfigKey, 2)
		j, _ := json.Marshal(&model.KVDoc{
			Key:   "reached_quota",
			Value: "1",
		})
		r, _ := http.NewRequest("POST", "/v1/test/kie/kv", bytes.NewBuffer(j))
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, chain)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	})
}
