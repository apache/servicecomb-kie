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

package kv_test

import (
	"time"

	"github.com/apache/servicecomb-kie/server/datasource"
	_ "github.com/apache/servicecomb-kie/test"

	"context"
	"testing"

	common2 "github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/kv"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
	"github.com/go-chassis/openlog"
	log "github.com/go-chassis/seclog"
	"github.com/stretchr/testify/assert"
)

var id string

func init() {
	log.Init(log.Config{
		Writers:     []string{"stdout"},
		LoggerLevel: "DEBUG",
	})

	logger := log.NewLogger("ut")
	openlog.SetLogger(logger)
}

func TestService_CreateOrUpdate(t *testing.T) {
	var err error
	err = session.Init(&datasource.Config{
		URI:     "mongodb://kie:123@127.0.0.1:27017/kie",
		Timeout: 10 * time.Second,
	})
	assert.NoError(t, err)
	kvsvc := &kv.Service{}
	t.Run("put kv timeout,with labels app and service", func(t *testing.T) {
		kv, err := kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app":     "mall",
				"service": "cart",
			},
			Domain:  "default",
			Project: "kv-test",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, kv.ID)
	})
	t.Run("put kv timeout,with labels app, service and version", func(t *testing.T) {
		kv, err := kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app":     "mall",
				"service": "cart",
				"version": "1.0.0",
			},
			Domain:  "default",
			Project: "kv-test",
		})
		oid, err := kvsvc.Exist(context.TODO(), "default", "timeout", "kv-test", datasource.WithLabels(map[string]string{
			"app":     "mall",
			"service": "cart",
			"version": "1.0.0",
		}))
		assert.NoError(t, err)
		assert.NotEmpty(t, kv.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, oid)
	})
	t.Run("put kv timeout,with labels app,and update value", func(t *testing.T) {
		beforeKV, err := kvsvc.Create(context.Background(), &model.KVDoc{
			Key:    "timeout",
			Value:  "1s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app": "mall",
			},
			Domain:  "default",
			Project: "kv-test",
		})
		assert.NoError(t, err)
		afterKV, err := kvsvc.Update(context.Background(), &model.UpdateKVRequest{
			ID:      beforeKV.ID,
			Value:   "3s",
			Domain:  "default",
			Project: "kv-test",
		})
		assert.Equal(t, "3s", afterKV.Value)
		savedKV, err := kvsvc.Exist(context.Background(), "default", "timeout", "kv-test", datasource.WithLabels(map[string]string{
			"app": "mall",
		}))
		assert.NoError(t, err)
		assert.Equal(t, afterKV.Value, savedKV.Value)
	})
}

func TestService_Create(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("create kv timeout,with labels app and service", func(t *testing.T) {
		result, err := kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app":     "mall",
				"service": "utCart",
			},
			Domain:  "default",
			Project: "kv-test",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, "2s", result.Value)
		id = result.ID
	})
	t.Run("create the same kv", func(t *testing.T) {
		_, err := kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app":     "mall",
				"service": "utCart",
			},
			Domain:  "default",
			Project: "kv-test",
		})
		assert.EqualError(t, err, datasource.ErrKVAlreadyExists.Error())
	})
	t.Run("list the kv", func(t *testing.T) {
		res, err := kvsvc.List(context.TODO(), "default", "kv-test", datasource.WithKey("wildcard(time*1)"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(res.Data))
		res, err = kvsvc.List(context.TODO(), "default", "kv-test", datasource.WithKey("wildcard(time*t)"))
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(res.Data))
	})
}

func TestService_Update(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("update kv by kvID", func(t *testing.T) {
		result, err := kvsvc.Update(context.TODO(), &model.UpdateKVRequest{
			ID:      id,
			Value:   "3s",
			Domain:  "default",
			Project: "kv-test",
		})
		assert.NoError(t, err)
		assert.Equal(t, "3s", result.Value)
	})
}

func TestService_Delete(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("delete kv by kvID", func(t *testing.T) {
		_, err := kvsvc.FindOneAndDelete(context.TODO(), id, "default", "kv-test")
		assert.NoError(t, err)
	})
}
