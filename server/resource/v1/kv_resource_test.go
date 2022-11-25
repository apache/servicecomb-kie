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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-kie/test"

	common2 "github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/plugin/qms"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/server/restful/restfultest"
	"github.com/stretchr/testify/assert"
)

var string32 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" //32
var string128 = string32 + string32 + string32 + string32
var string1024 = string128 + string128 + string128 + string128 + string128 + string128 + string128 + string128
var string8192 = string1024 + string1024 + string1024 + string1024 + string1024 + string1024 + string1024 + string1024
var string65536 = string8192 + string8192 + string8192 + string8192 + string8192 + string8192 + string8192 + string8192
var string131072 = string65536 + string65536

func TestKVResource_Post(t *testing.T) {
	t.Run("post kv, label is invalid, should return err", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:    "timeout",
			Value:  "1s",
			Labels: map[string]string{"service": strings.Repeat("x", 161)},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusBadRequest, resp.Result().StatusCode)
	})
	t.Run("post kv, label is service", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:    "timeout",
			Value:  "1s",
			Labels: map[string]string{"service": "utService"},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
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
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
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
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.KVDoc{}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.NotEmpty(t, data.ID)
	})
	t.Run("post kv, value type is xml, should success", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:       "xml",
			Value:     "a",
			ValueType: "xml",
			Labels:    map[string]string{"a": "a"},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
	t.Run("post kv, labels is nil, should success", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:   "no_labels",
			Value: "without_labels",
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
	t.Run("post kv, has one label, key of label is a empty string, should return err", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:    "withoutKeyOfLabels",
			Value:  "withoutKeyOfLabels",
			Labels: map[string]string{"": "a"},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusBadRequest, resp.Result().StatusCode)
	})
	t.Run("post kv, has one label, value of label is a empty string, should success", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:    "withoutValueOfLabels",
			Value:  "withoutValueOfLabels",
			Labels: map[string]string{"a": ""},
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.KVDoc{}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(data.Labels["a"]))
	})
	t.Run("post kv, length of value is 131072, should success", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:   "postMaxValueOfKie",
			Value: string131072,
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.KVDoc{}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, 131072, len(data.Value))
	})
	t.Run("post kv, length of value is 131073, should return err", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:   "postGreaterThanMaxValueOfKie",
			Value: string131072 + "a",
		}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusBadRequest, resp.Result().StatusCode)
	})
}
func TestKVResource_List(t *testing.T) {
	t.Run("list kv by service label, should return 3 kvs", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code, string(body))
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.Data))
	})
	var rev string
	t.Run("list kv by service label, exact match,should return 2 kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&match=exact", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
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
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&"+common2.QueryParamRev+"="+rev, nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotModified, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with old rev param,should return latest revision", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&"+common2.QueryParamRev+"=1", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with wait and old rev param,should return latest revision,no wait", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&wait=1s&"+common2.QueryParamRev+"=1", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		start := time.Now()
		c.ServeHTTP(resp, r)
		duration := time.Since(start)
		t.Log(duration)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with wait and current rev param,should wait and return 304 ", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&"+common2.QueryParamRev+"=1", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
		rev = resp.Header().Get(common2.HeaderRevision)
		t.Log(rev)

		r2, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&wait=2s&"+common2.QueryParamRev+"="+rev, nil)
		r2.Header.Set("Content-Type", "application/json")
		t.Log(r2.URL.String())
		resp2 := httptest.NewRecorder()
		c.ServeHTTP(resp2, r2)
		rev = resp2.Header().Get(common2.HeaderRevision)
		t.Log(rev)
		body, err := ioutil.ReadAll(resp2.Body)
		time.Sleep(1 * time.Second)
		t.Log(string(body))
		assert.Equal(t, http.StatusNotModified, resp2.Result().StatusCode)
	})
	t.Run("list kv by service label, with wait and larger rev param,should return latest revision,no wait", func(t *testing.T) {
		revNum, _ := strconv.ParseInt(rev, 10, 64)
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&wait=1s&"+common2.QueryParamRev+fmt.Sprintf("=%d", revNum+100), nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		start := time.Now()
		c.ServeHTTP(resp, r)
		duration := time.Since(start)
		t.Log(duration)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
	t.Run("list kv by service label, with wait param,will exceed 1s and return 304", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&wait=1s", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		start := time.Now()
		c.ServeHTTP(resp, r)
		duration := time.Since(start)
		t.Log(duration)
	})
	t.Run("list kv by service label offset, should return 1kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=service:utService&offset=0&limit=1", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
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
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?label=match:test&wait=10s&match=exact", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
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
			r2, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
			r2.Header.Set("Content-Type", "application/json")
			kvr2 := &v1.KVResource{}
			c2, _ := restfultest.New(kvr2, nil)
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
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?key=timeout&label=service:utService&match=exact", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
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
	t.Run("get one key, fuzzy match,should return 2 kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?key=beginWith(TIME)", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result.Data))

		r, _ = http.NewRequest("GET", "/v1/kv_test/kie/kv?key=beginWith(IME)", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr = &v1.KVResource{}
		c, err = restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp = httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err = ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result = &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(result.Data))
	})
	t.Run("get one key, wildcard match,should return 2 kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?key=wildcard(*IME*)", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result.Data))

		r, _ = http.NewRequest("GET", "/v1/kv_test/kie/kv?key=wildcard(TIME)", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr = &v1.KVResource{}
		c, err = restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp = httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err = ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result = &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(result.Data))
	})
	t.Run("get one key by service label should return 2 kv,delete one", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv?key=timeout&label=service:utService", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result.Data))

		r2, _ := http.NewRequest("DELETE", "/v1/kv_test/kie/kv/"+result.Data[0].ID, nil)
		c2, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp2 := httptest.NewRecorder()
		c2.ServeHTTP(resp2, r2)
		assert.Equal(t, http.StatusNoContent, resp2.Code)

	})
}
func TestKVResource_Upload(t *testing.T) {
	t.Run("test force with the same key and the same labels, and one invalid input, should return 2 success and 1 failure", func(t *testing.T) {
		input := new(v1.KVUploadBody)
		input.Data = []*model.KVDoc{
			{
				Key:    "1",
				Value:  "1",
				Labels: map[string]string{"2": "2"},
			},
			{
				Key:    "1",
				Value:  "1",
				Status: "invalid",
				Labels: map[string]string{"1": "1"},
			},
			{
				Key:    "1",
				Value:  "1-update",
				Labels: map[string]string{"2": "2"},
			},
		}
		j, _ := json.Marshal(input)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/file?override=force", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.DocRespOfUpload{
			Success: []*model.KVDoc{},
			Failure: []*model.DocFailedOfUpload{},
		}
		fmt.Println("test=======================================================")
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 1, len(data.Failure))
		assert.Equal(t, 2, len(data.Success))
		assert.Equal(t, data.Success[0].ID, data.Success[1].ID)
		assert.Equal(t, "1-update", data.Success[1].Value)
	})
	t.Run("test force with the same key and not the same labels and ont invalid input, should return 2 success and 1 failure", func(t *testing.T) {
		input := new(v1.KVUploadBody)
		input.Data = []*model.KVDoc{
			{
				Key:    "2",
				Value:  "2",
				Labels: map[string]string{"1": "1"},
			},
			{
				Key:    "2",
				Value:  "2",
				Status: "invalid",
				Labels: map[string]string{"1": "1"},
			},

			{
				Key:    "2",
				Value:  "2",
				Labels: map[string]string{"2": "2"},
			},
		}
		j, _ := json.Marshal(input)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/file?override=force", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.DocRespOfUpload{
			Success: []*model.KVDoc{},
			Failure: []*model.DocFailedOfUpload{},
		}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 1, len(data.Failure))
		assert.Equal(t, 2, len(data.Success))
		assert.NotEqual(t, data.Success[0].ID, data.Success[1].ID)
	})
	t.Run("test skip, with one invalid input, should return 2 success and 2 failure", func(t *testing.T) {
		input := new(v1.KVUploadBody)
		input.Data = []*model.KVDoc{
			{
				Key:    "3",
				Value:  "1",
				Labels: map[string]string{"2": "2"},
			},
			{
				Key:    "2",
				Value:  "2",
				Status: "invalid",
				Labels: map[string]string{"1": "1"},
			},
			{
				Key:    "3",
				Value:  "1-update",
				Labels: map[string]string{"2": "2"},
			},
			{
				Key:       "4",
				Value:     "1",
				Labels:    map[string]string{"2": "2"},
				ValueType: "text",
				Status:    "enabled",
			},
		}
		j, _ := json.Marshal(input)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/file?override=skip", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.DocRespOfUpload{
			Success: []*model.KVDoc{},
			Failure: []*model.DocFailedOfUpload{},
		}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 2, len(data.Failure))
		assert.Equal(t, 2, len(data.Success))
		assert.Equal(t, "1", data.Success[0].Value)
		assert.Equal(t, "1", data.Success[1].Value)
		assert.Equal(t, "validate failed, field: KVDoc.Status, rule: ^$|^(enabled|disabled)$", data.Failure[0].ErrMsg)
		assert.Equal(t, "skip overriding duplicate kvs", data.Failure[1].ErrMsg)
	})
	t.Run("test abort, with one invalid input, should return 1 success and 3 failure", func(t *testing.T) {
		input := new(v1.KVUploadBody)
		input.Data = []*model.KVDoc{
			{
				Key:    "5",
				Value:  "2",
				Labels: map[string]string{"1": "1"},
			},
			{
				Key:    "5",
				Value:  "2-update",
				Labels: map[string]string{"1": "1"},
			},
			{
				Key:    "5",
				Value:  "2-update",
				Status: "invalid",
				Labels: map[string]string{"1": "1"},
			},
			{
				Key:    "6",
				Value:  "2",
				Labels: map[string]string{"4": "4"},
			},
		}
		j, _ := json.Marshal(input)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/file?override=abort", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.DocRespOfUpload{
			Success: []*model.KVDoc{},
			Failure: []*model.DocFailedOfUpload{},
		}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 3, len(data.Failure))
		assert.Equal(t, 1, len(data.Success))
		assert.Equal(t, "2", data.Success[0].Value)
	})
	t.Run("test upload 2 configurations, the value of one of the two configurations is 128kb, should return 2 success", func(t *testing.T) {
		input := new(v1.KVUploadBody)
		input.Data = []*model.KVDoc{
			{
				Key:    "128kb",
				Value:  string131072,
				Labels: map[string]string{"2": "2"},
			},
			{
				Key:    "64kb",
				Value:  string65536,
				Labels: map[string]string{"1": "1"},
			},
		}
		j, _ := json.Marshal(input)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/file?override=force", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.DocRespOfUpload{
			Success: []*model.KVDoc{},
			Failure: []*model.DocFailedOfUpload{},
		}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 0, len(data.Failure))
		assert.Equal(t, 2, len(data.Success))
		assert.Equal(t, 131072, len(data.Success[0].Value))
		assert.Equal(t, 65536, len(data.Success[1].Value))
	})
	t.Run("test upload 2 configurations, one of the values is greater than 128kb, should return 1 success and 1 failure", func(t *testing.T) {
		input := new(v1.KVUploadBody)
		input.Data = []*model.KVDoc{
			{
				Key:    "uploadGreaterThanMaxValueOfKie",
				Value:  string131072 + "a",
				Labels: map[string]string{"2": "2"},
			},
			{
				Key:    "1kb",
				Value:  string1024,
				Labels: map[string]string{"1": "1"},
			},
		}
		j, _ := json.Marshal(input)
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/file?override=force", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)

		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.DocRespOfUpload{
			Success: []*model.KVDoc{},
			Failure: []*model.DocFailedOfUpload{},
		}
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 1, len(data.Failure))
		assert.Equal(t, 1, len(data.Success))
		assert.Equal(t, "uploadGreaterThanMaxValueOfKie", data.Failure[0].Key)
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
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
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
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv/"+id, nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
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
	kvUpdate := &model.UpdateKVRequest{
		Value: "admin",
	}
	t.Run("update the kv, set the value of user to admin", func(t *testing.T) {
		j, _ := json.Marshal(kvUpdate)
		r, _ := http.NewRequest("PUT", "/v1/kv_test/kie/kv/"+id, bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
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
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv/"+id, nil)

		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
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

		err := archaius.Set(qms.QuotaConfigKey, 2)
		assert.NoError(t, err)
		j, _ := json.Marshal(&model.KVDoc{
			Key:   "reached_quota",
			Value: "1",
		})
		r, _ := http.NewRequest("POST", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, _ := restfultest.New(kvr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	})
}
func TestKVResource_DeleteList(t *testing.T) {
	var ids []string
	t.Run("get ids of all kvs", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(result.Data))
		for _, kv := range result.Data {
			ids = append(ids, kv.ID)
		}
	})
	t.Run("delete all kvs by ids", func(t *testing.T) {
		j, _ := json.Marshal(v1.DeleteBody{IDs: ids})
		r, _ := http.NewRequest("DELETE", "/v1/kv_test/kie/kv", bytes.NewBuffer(j))

		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		assert.Equal(t, http.StatusNoContent, resp.Code)

	})
	t.Run("get all kvs again, should return 0 kv", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/v1/kv_test/kie/kv", nil)
		r.Header.Set("Content-Type", "application/json")
		kvr := &v1.KVResource{}
		c, err := restfultest.New(kvr, nil)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		result := &model.KVResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(result.Data))
	})
}
