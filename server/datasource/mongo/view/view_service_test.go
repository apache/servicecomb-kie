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

package view_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	common2 "github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/kv"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/view"
	_ "github.com/apache/servicecomb-kie/test"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestGet(t *testing.T) {
	var err error
	err = session.Init(&datasource.Config{
		URI:     "mongodb://kie:123@127.0.0.1:27017/kie",
		Timeout: 10 * time.Second,
	})
	assert.NoError(t, err)
	kvsvc := &kv.Service{}
	t.Run("put view data", func(t *testing.T) {
		kv, err := kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app":     "mall",
				"service": "cart",
				"view":    "view_test",
			},
			Domain:  "default",
			Project: "view_test",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, kv.ID)

		kv, err = kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "timeout",
			Value:  "2s",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app": "mall",
			},
			Domain:  "default",
			Project: "view_test",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, kv.ID)

		kv, err = kvsvc.Create(context.TODO(), &model.KVDoc{
			Key:    "retry",
			Value:  "2",
			Status: common2.StatusEnabled,
			Labels: map[string]string{
				"app": "mall",
			},
			Domain:  "default",
			Project: "view_test",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, kv.ID)
	})

	svc := &view.Service{}
	t.Run("create and get view content", func(t *testing.T) {
		view1, err := svc.Create(context.TODO(), &model.ViewDoc{
			Display: "timeout_config",
			Project: "view_test",
			Domain:  "default",
		}, datasource.WithKey("timeout"))
		assert.NoError(t, err)
		assert.NotEmpty(t, view1.ID)
		view2, err := svc.Create(context.TODO(), &model.ViewDoc{
			Display: "mall_config",
			Project: "view_test",
			Domain:  "default",
		}, datasource.WithLabels(map[string]string{
			"app": "mall",
		}))
		assert.NoError(t, err)
		assert.NotEmpty(t, view2.ID)

		resp1, err := svc.GetContent(context.TODO(), view1.ID, "default", "view_test")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp1.Data))
		assert.Equal(t, "timeout", resp1.Data[0].Key)

		resp2, err := svc.GetContent(context.TODO(), view2.ID, "default", "view_test")
		assert.NoError(t, err)
		assert.Equal(t, "mall", resp1.Data[0].Labels["app"])
		t.Log(resp2.Data)
	})
	t.Run(" list view", func(t *testing.T) {
		r, err := svc.List(context.TODO(), "default", "view_test")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(r.Data))
	})

}

func TestService_List(t *testing.T) {
	var pipeline mongo.Pipeline = []bson.D{
		{{
			"$match",
			bson.D{{"domain", "default"}, {"project", "default"}},
		}},
	}

	s, err := json.Marshal(pipeline)
	assert.NoError(t, err)
	t.Log(string(s))
}
