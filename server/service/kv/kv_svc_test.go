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
	"github.com/apache/servicecomb-kie/server/datasource"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
	_ "github.com/apache/servicecomb-kie/test"
	"github.com/go-chassis/cari/config"

	"context"
	"testing"

	common2 "github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/openlog"
	log "github.com/go-chassis/seclog"
	"github.com/stretchr/testify/assert"
)

var project = "kv-test"
var domain = "default"
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
	t.Run("put kv timeout,with labels app and service", func(t *testing.T) {
		kv, err := kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app":     "mall",
				"service": "cart",
			},
			Domain:  domain,
			Project: project,
		})
		assert.Nil(t, err)
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
			Domain:  domain,
			Project: project,
		})
		oid, err2 := kvsvc.Get(context.TODO(), &model.GetKVRequest{
			Domain:  domain,
			Project: project,
			ID:      kv.ID,
		})
		assert.NoError(t, err2)
		assert.NotEmpty(t, kv.ID)
		assert.Nil(t, err)
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
			Domain:  domain,
			Project: project,
		})
		assert.Nil(t, err)
		_, err2 := kvsvc.Update(context.Background(), &model.UpdateKVRequest{
			ID:      beforeKV.ID,
			Value:   "3s",
			Domain:  domain,
			Project: project,
		})
		assert.NoError(t, err2)
		savedKV, err2 := kvsvc.Get(context.TODO(), &model.GetKVRequest{
			Domain:  domain,
			Project: project,
			ID:      beforeKV.ID,
		})
		assert.NoError(t, err2)
		assert.Equal(t, "3s", savedKV.Value)
	})
}

func TestService_Create(t *testing.T) {
	t.Run("create kv timeout,with labels app and service", func(t *testing.T) {
		result, err := kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app":     "mall",
				"service": "utCart",
			},
			Domain:  domain,
			Project: project,
		})
		assert.Nil(t, err)
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
			Domain:  domain,
			Project: project,
		})
		assert.EqualError(t, err,
			config.NewError(config.ErrRecordAlreadyExists, datasource.ErrKVAlreadyExists.Error()).Error())
	})
	t.Run("list the kv", func(t *testing.T) {
		res, err := kvsvc.List(context.TODO(), project, domain,
			datasource.WithKey("wildcard(time*1)"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(res.Data))
		res, err = kvsvc.List(context.TODO(), project, domain,
			datasource.WithKey("wildcard(time*t)"))
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(res.Data))
	})
}

func TestService_Update(t *testing.T) {
	t.Run("update kv by kvID", func(t *testing.T) {
		_, err := kvsvc.Update(context.TODO(), &model.UpdateKVRequest{
			ID:      id,
			Value:   "3s",
			Domain:  domain,
			Project: project,
		})
		assert.NoError(t, err)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("delete kv by kvID", func(t *testing.T) {
		_, err := kvsvc.FindOneAndDelete(context.TODO(), id, project, domain)
		assert.NoError(t, err)
	})
}
