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

	_ "github.com/apache/servicecomb-kie/test"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
)

func TestGetHistory(t *testing.T) {
	ctx := context.TODO()
	kv, err := kvsvc.Create(ctx, &model.KVDoc{
		Key:    "TestGetHistory",
		Value:  "2s",
		Status: common.StatusEnabled,
		Labels: map[string]string{
			"app":     "mall",
			"service": "cart",
		},
		Domain:  "default",
		Project: "kv-test",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, kv.ID)

	_, uErr := kvsvc.Update(ctx, &model.UpdateKVRequest{
		ID:      kv.ID,
		Value:   "3s",
		Domain:  "default",
		Project: "kv-test",
	})
	assert.NoError(t, uErr)

	t.Run("after create kv, should has history", func(t *testing.T) {
		h, err := datasource.GetBroker().GetHistoryDao().GetHistory(ctx, kv.ID, "kv-test", "default")
		assert.NoError(t, err)
		assert.Equal(t, 2, h.Total)
		assert.Equal(t, 2, len(h.Data))
	})

	t.Run("test paging, should pass", func(t *testing.T) {
		resp, err := datasource.GetBroker().GetHistoryDao().GetHistory(ctx, kv.ID, "kv-test", "default",
			datasource.WithOffset(0), datasource.WithLimit(1))
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, 1, len(resp.Data))
		assert.Equal(t, "3s", resp.Data[0].Value)

		resp, err = datasource.GetBroker().GetHistoryDao().GetHistory(ctx, kv.ID, "kv-test", "default",
			datasource.WithOffset(1), datasource.WithLimit(1))
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, 1, len(resp.Data))
		assert.Equal(t, "2s", resp.Data[0].Value)
	})

	_, delErr := kvsvc.FindOneAndDelete(ctx, kv.ID, "kv-test", "default")
	assert.NoError(t, delErr)
}
