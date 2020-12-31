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
	"context"
	common2 "github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/kv"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-chassis/openlog"
	log "github.com/go-chassis/seclog"
	"github.com/stretchr/testify/assert"
	"testing"
)

var id, id2 string

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
	config.Configurations = &config.Config{DB: config.DB{URI: "mongodb://kie:123@127.0.0.1:27017/kie"}}
	err = session.Init()
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
		oid, err := kvsvc.Exist(context.TODO(), "default", "timeout", "kv-test", service.WithLabels(map[string]string{
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
		savedKV, err := kvsvc.Exist(context.Background(), "default", "timeout", "kv-test", service.WithLabels(map[string]string{
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
		assert.EqualError(t, err, session.ErrKVAlreadyExists.Error())
	})
}

func TestService_CreateList(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("create a kv timeout,with labels app and service", func(t *testing.T) {
		result, err := kvsvc.CreateList(context.TODO(), &model.KVListDoc{
			KVListDoc: []*model.KVDoc{
				{
					Key:    "timeout",
					Value:  "2s",
					Status: common2.StatusEnabled,
					Labels: map[string]string{
						"app":     "mall",
						"service": "utCart",
						"version": "1.0.1",
					},
					Domain:  "default",
					Project: "kv-test",
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result.KVListDoc[0].ID)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "2s", result.KVListDoc[0].Value)
		id2 = result.KVListDoc[0].ID

	})
	t.Run("create a kv list which contains a exist kv value", func(t *testing.T) {
		_, err := kvsvc.CreateList(context.TODO(), &model.KVListDoc{
			KVListDoc: []*model.KVDoc{
				{
					Key:    "timeout",
					Value:  "2s",
					Status: common2.StatusEnabled,
					Labels: map[string]string{
						"app":     "mall",
						"service": "utCart",
						"version": "1.0.1",
					},
					Domain:  "default",
					Project: "kv-test",
				},
				{
					Key:    "test",
					Value:  "5s",
					Status: common2.StatusEnabled,
					Labels: map[string]string{
						"app":     "mall",
						"service": "utCart",
					},
					Domain:  "default",
					Project: "kv-test",
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("create a kv list with more than one kv value", func(t *testing.T) {
		_, err := kvsvc.CreateList(context.TODO(), &model.KVListDoc{
			KVListDoc: []*model.KVDoc{
				{
					Key:    "test1",
					Value:  "2s",
					Status: common2.StatusEnabled,
					Labels: map[string]string{
						"app":     "mall",
						"service": "utCart",
						"version": "1.0.1",
					},
					Domain:  "default",
					Project: "kv-test",
				},
				{
					Key:    "test2",
					Value:  "5s",
					Status: common2.StatusEnabled,
					Labels: map[string]string{
						"app":     "mall",
						"service": "utCart",
					},
					Domain:  "default",
					Project: "kv-test",
				},
			},
		})
		assert.NoError(t, err)
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

func TestService_UpdateList(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("update kv by kvID", func(t *testing.T) {
		result, err := kvsvc.UpdateList(context.TODO(), &model.UpdateKVListRequest{
			UpdateKVList: []*model.UpdateKVRequest{
				{
					ID:      id,
					Value:   "5s",
					Domain:  "default",
					Project: "kv-test",
				},
				{
					ID:      id2,
					Value:   "8s",
					Domain:  "default",
					Project: "kv-test",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, "5s", result.KVListDoc[0].Value)
		assert.Equal(t, "8s", result.KVListDoc[1].Value)
	})
}

func TestService_Delete(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("delete kv by kvID", func(t *testing.T) {
		_, err := kvsvc.FindOneAndDelete(context.TODO(), id, "default", "kv-test")
		assert.NoError(t, err)
	})
}
