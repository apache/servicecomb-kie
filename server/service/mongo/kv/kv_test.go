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
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/kv"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_CreateOrUpdate(t *testing.T) {
	var err error
	config.Configurations = &config.Config{DB: config.DB{URI: "mongodb://kie:123@127.0.0.1:27017/kie"}}
	err = session.Init()
	assert.NoError(t, err)
	kvsvc := &kv.Service{}
	t.Run("put kv timeout,with labels app and service", func(t *testing.T) {
		kv, err := kvsvc.CreateOrUpdate(context.TODO(), &model.KVDoc{
			Key:   "timeout",
			Value: "2s",
			Labels: map[string]string{
				"app":     "mall",
				"service": "cart",
			},
			Domain:  "default",
			Project: "test",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, kv.ID)
	})
	t.Run("put kv timeout,with labels app, service and version", func(t *testing.T) {
		kv, err := kvsvc.CreateOrUpdate(context.TODO(), &model.KVDoc{
			Key:   "timeout",
			Value: "2s",
			Labels: map[string]string{
				"app":     "mall",
				"service": "cart",
				"version": "1.0.0",
			},
			Domain:  "default",
			Project: "test",
		})
		oid, err := kvsvc.Exist(context.TODO(), "default", "timeout", "test", service.WithLabels(map[string]string{
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
		beforeKV, err := kvsvc.CreateOrUpdate(context.Background(), &model.KVDoc{
			Key:   "timeout",
			Value: "1s",
			Labels: map[string]string{
				"app": "mall",
			},
			Domain:  "default",
			Project: "test",
		})
		assert.NoError(t, err)
		kvs1, err := kvsvc.FindKV(context.Background(), "default", "test",
			service.WithKey("timeout"),
			service.WithLabels(map[string]string{
				"app": "mall",
			}),
			service.WithExactLabels())
		assert.Equal(t, beforeKV.Value, kvs1[0].Data[0].Value)
		afterKV, err := kvsvc.CreateOrUpdate(context.Background(), &model.KVDoc{
			Key:   "timeout",
			Value: "3s",
			Labels: map[string]string{
				"app": "mall",
			},
			Domain:  "default",
			Project: "test",
		})
		assert.Equal(t, beforeKV.ID, afterKV.ID)
		savedKV, err := kvsvc.Exist(context.Background(), "default", "timeout", "test", service.WithLabels(map[string]string{
			"app": "mall",
		}))
		assert.NoError(t, err)
		assert.Equal(t, beforeKV.ID, savedKV.ID)
		kvs, err := kvsvc.FindKV(context.Background(), "default", "test",
			service.WithKey("timeout"),
			service.WithLabels(map[string]string{
				"app": "mall",
			}),
			service.WithExactLabels())
		assert.Equal(t, afterKV.Value, kvs[0].Data[0].Value)
	})

}

func TestService_FindKV(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("exact find by kv and labels with label app", func(t *testing.T) {
		kvs, err := kvsvc.FindKV(context.Background(), "default", "test",
			service.WithKey("timeout"),
			service.WithLabels(map[string]string{
				"app": "mall",
			}),
			service.WithExactLabels())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(kvs))
	})
	t.Run("greedy find by labels,with labels app ans service ", func(t *testing.T) {
		kvs, err := kvsvc.FindKV(context.Background(), "default", "test",
			service.WithLabels(map[string]string{
				"app":     "mall",
				"service": "cart",
			}))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(kvs))
	})
}
func TestService_Delete(t *testing.T) {
	kvsvc := &kv.Service{}
	t.Run("delete key by kvID", func(t *testing.T) {
		kv1, err := kvsvc.CreateOrUpdate(context.Background(), &model.KVDoc{
			Key:   "timeout",
			Value: "20s",
			Labels: map[string]string{
				"env": "test",
			},
			Domain:  "default",
			Project: "test",
		})
		assert.NoError(t, err)

		err = kvsvc.Delete(context.TODO(), kv1.ID, "default", "test")
		assert.NoError(t, err)

	})
	t.Run("miss id", func(t *testing.T) {
		err := kvsvc.Delete(context.TODO(), "", "default", "test")
		assert.Error(t, err)
	})
	t.Run("miss domain", func(t *testing.T) {
		err := kvsvc.Delete(context.TODO(), "2", "", "test")
		assert.Equal(t, session.ErrMissingDomain, err)
	})
}
