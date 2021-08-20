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

package datasource_test

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	ctx := context.TODO()
	kv1, err := kvsvc.Create(ctx, &model.KVDoc{
		Key:    "TestList1",
		Value:  "2s",
		Status: common.StatusEnabled,
		Labels: map[string]string{
			"app":     "mall",
			"service": "cart",
		},
		Domain:  "default",
		Project: "kv-list-test",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, kv1.ID)
	defer kvsvc.FindOneAndDelete(ctx, kv1.ID, "kv-list-test", "default")

	kv2, err := kvsvc.Create(ctx, &model.KVDoc{
		Key:    "TestList2",
		Value:  "3s",
		Status: common.StatusEnabled,
		Labels: map[string]string{
			"app":     "mall",
			"service": "cart",
		},
		Domain:  "default",
		Project: "kv-list-test",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, kv2.ID)
	defer kvsvc.FindOneAndDelete(ctx, kv2.ID, "kv-list-test", "default")

	kv3, err := kvsvc.Create(ctx, &model.KVDoc{
		Key:    "TestList3",
		Value:  "4s",
		Status: common.StatusEnabled,
		Labels: map[string]string{
			"app":     "mall",
			"service": "cart",
		},
		Domain:  "default",
		Project: "kv-list-test",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, kv3.ID)
	defer kvsvc.FindOneAndDelete(ctx, kv3.ID, "kv-list-test", "default")

	t.Run("after create kv, should list results", func(t *testing.T) {
		h, err := datasource.GetBroker().GetKVDao().List(ctx, "kv-list-test", "default")
		assert.NoError(t, err)
		assert.Equal(t, 3, h.Total)
		assert.Equal(t, 3, len(h.Data))
	})

	t.Run("test paging, should pass", func(t *testing.T) {
		resp, err := datasource.GetBroker().GetKVDao().List(ctx, "kv-list-test", "default",
			datasource.WithOffset(0), datasource.WithLimit(2))
		assert.NoError(t, err)
		assert.Equal(t, 3, resp.Total)
		assert.Equal(t, 2, len(resp.Data))

		resp, err = datasource.GetBroker().GetKVDao().List(ctx, "kv-list-test", "default",
			datasource.WithOffset(2), datasource.WithLimit(2))
		assert.NoError(t, err)
		assert.Equal(t, 3, resp.Total)
		assert.Equal(t, 1, len(resp.Data))
	})
}
