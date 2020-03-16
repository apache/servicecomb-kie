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
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	handler2 "github.com/apache/servicecomb-kie/server/handler"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/server/restful/restfultest"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/apache/servicecomb-kie/server/service/mongo"
)

func TestHistoryResource_GetRevisions(t *testing.T) {
	kv := &model.KVDoc{
		Key:   "test",
		Value: "revisions",
		Labels: map[string]string{
			"test": "revisions",
		},
		Domain:  "default",
		Project: "test",
	}
	kv, _ = service.KVService.CreateOrUpdate(context.Background(), kv)
	path := fmt.Sprintf("/v1/test/kie/revision/%s", kv.ID)
	r, _ := http.NewRequest("GET", path, nil)
	revision := &v1.HistoryResource{}
	chain, _ := handler.GetChain(common.Provider, "")
	c, err := restfultest.New(revision, chain)
	assert.NoError(t, err)
	resp := httptest.NewRecorder()
	c.ServeHTTP(resp, r)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	data := make([]*model.KVDoc, 0)
	err = json.Unmarshal(body, &data)
	assert.NoError(t, err)
	before := len(data)
	assert.GreaterOrEqual(t, before, 1)

	t.Run("put again, should has 2 revision", func(t *testing.T) {
		kv.Domain = "default"
		kv.Project = "test"
		kv, err = service.KVService.CreateOrUpdate(context.Background(), kv)
		assert.NoError(t, err)
		path := fmt.Sprintf("/v1/test/kie/revision/%s", kv.ID)
		r, _ := http.NewRequest("GET", path, nil)
		revision := &v1.HistoryResource{}
		chain, _ := handler.GetChain(common.Provider, "")
		c, err := restfultest.New(revision, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := make([]*model.KVDoc, 0)
		err = json.Unmarshal(body, &data)
		assert.Equal(t, before+1, len(data))
	})

}

func TestHistoryResource_GetPollingData(t *testing.T) {
	t.Run("list kv by service label, to create a polling data", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/kv", nil)
		noopH := &handler2.NoopAuthHandler{}
		noopH2 := &handler2.TrackHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain3", noopH.Name(), noopH2.Name())
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Session-Id", "test")
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
	})
	t.Run("get polling data", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/test/kie/track?sessionId=test", nil)
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain3", noopH.Name())
		r.Header.Set("Content-Type", "application/json")
		revision := &v1.HistoryResource{}
		c, err := restfultest.New(revision, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.PollingDataResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result.Data)
	})

}

func Test_HeathCheck(t *testing.T) {
	path := fmt.Sprintf("/v1/health")
	r, _ := http.NewRequest("GET", path, nil)
	noopH := &handler2.NoopAuthHandler{}
	revision := &v1.HistoryResource{}
	chain, _ := handler.CreateChain(common.Provider, "", noopH.Name())
	c, err := restfultest.New(revision, chain)
	assert.NoError(t, err)
	resp := httptest.NewRecorder()
	c.ServeHTTP(resp, r)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	data := &model.DocHealthCheck{}
	err = json.Unmarshal(body, &data)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
