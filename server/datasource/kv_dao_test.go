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

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
	"github.com/apache/servicecomb-kie/test"
	emodel "github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
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

	// delete
	_, delErr := kvsvc.FindOneAndDelete(ctx, kv1.ID, "kv-list-test", "default")
	assert.NoError(t, delErr)
	_, delErr = kvsvc.FindOneAndDelete(ctx, kv2.ID, "kv-list-test", "default")
	assert.NoError(t, delErr)
	_, delErr = kvsvc.FindOneAndDelete(ctx, kv3.ID, "kv-list-test", "default")
	assert.NoError(t, delErr)
}

func TestWithSync(t *testing.T) {
	if test.IsEmbeddedetcdMode() {
		return
	}

	t.Run("create kv with sync enabled", func(t *testing.T) {
		t.Run("creating a kv will create a task should pass", func(t *testing.T) {
			// set the sync enabled
			ctx := context.WithValue(context.Background(), "sync", true)
			kv1, err := kvsvc.Create(ctx, &model.KVDoc{
				Key:    "sync-create",
				Value:  "2s",
				Status: common.StatusEnabled,
				Labels: map[string]string{
					"app":     "sync-create",
					"service": "sync-create",
				},
				Domain:  "default",
				Project: "sync-create",
			})
			assert.Nil(t, err)
			assert.NotEmpty(t, kv1.ID)

			listReq := emodel.ListTaskRequest{
				Domain:  "default",
				Project: "sync-create",
			}
			_, tempErr := kvsvc.FindOneAndDelete(ctx, kv1.ID, "sync-create", "default")
			assert.Nil(t, tempErr)
			resp, tempErr := kvsvc.List(ctx, "sync-create", "default")
			assert.Nil(t, tempErr)
			assert.Equal(t, 0, resp.Total)
			tasks, tempErr := task.List(ctx, &listReq)
			assert.Nil(t, tempErr)
			assert.Equal(t, 2, len(tasks))
			tempErr = task.Delete(ctx, tasks...)
			assert.Nil(t, tempErr)
			tbListReq := emodel.ListTombstoneRequest{
				Domain:       "default",
				Project:      "sync-create",
				ResourceType: datasource.ConfigResource,
			}
			tombstones, tempErr := tombstone.List(ctx, &tbListReq)
			assert.Equal(t, 1, len(tombstones))
			tempErr = tombstone.Delete(ctx, tombstones...)
			assert.Nil(t, tempErr)
		})
	})

	t.Run("update kv with sync enabled", func(t *testing.T) {
		t.Run("creating two kvs and updating them will create four tasks should pass", func(t *testing.T) {
			// set the sync enabled
			ctx := context.WithValue(context.Background(), "sync", true)
			kv1, err := kvsvc.Create(ctx, &model.KVDoc{
				Key:    "sync-update-one",
				Value:  "2s",
				Status: common.StatusEnabled,
				Labels: map[string]string{
					"app":     "sync-update",
					"service": "sync-update",
				},
				Domain:  "default",
				Project: "sync-update",
			})
			assert.Nil(t, err)
			assert.NotEmpty(t, kv1.ID)
			kv2, err := kvsvc.Create(ctx, &model.KVDoc{
				Key:    "sync-update-two",
				Value:  "2s",
				Status: common.StatusEnabled,
				Labels: map[string]string{
					"app":     "sync-update",
					"service": "sync-update",
				},
				Domain:  "default",
				Project: "sync-update",
			})
			assert.Nil(t, err)
			assert.NotEmpty(t, kv2.ID)
			kv1, tmpErr := kvsvc.Update(ctx, &model.UpdateKVRequest{
				ID:      kv1.ID,
				Value:   "3s",
				Domain:  "default",
				Project: "sync-update",
			})
			assert.Nil(t, tmpErr)
			assert.NotEmpty(t, kv1.ID)
			kv2, tmpErr = kvsvc.Update(ctx, &model.UpdateKVRequest{
				ID:      kv2.ID,
				Value:   "3s",
				Domain:  "default",
				Project: "sync-update",
			})
			assert.Nil(t, tmpErr)
			assert.NotEmpty(t, kv2.ID)
			_, tempErr := kvsvc.FindManyAndDelete(ctx, []string{kv1.ID, kv2.ID}, "sync-update", "default")
			assert.Nil(t, tempErr)
			resp, tempErr := kvsvc.List(ctx, "sync-update", "default")
			assert.Nil(t, tempErr)
			assert.Equal(t, 0, resp.Total)
			listReq := emodel.ListTaskRequest{
				Domain:  "default",
				Project: "sync-update",
			}
			tasks, tempErr := task.List(ctx, &listReq)
			assert.Nil(t, tempErr)
			assert.Equal(t, 6, len(tasks))
			tempErr = task.Delete(ctx, tasks...)
			assert.Nil(t, tempErr)
			tbListReq := emodel.ListTombstoneRequest{
				Domain:       "default",
				Project:      "sync-update",
				ResourceType: datasource.ConfigResource,
			}
			tombstones, tempErr := tombstone.List(ctx, &tbListReq)
			assert.Equal(t, 2, len(tombstones))
			tempErr = tombstone.Delete(ctx, tombstones...)
			assert.Nil(t, tempErr)
		})
	})

}
