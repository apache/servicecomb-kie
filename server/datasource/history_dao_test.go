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
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
	"testing"

	common2 "github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"

	_ "github.com/apache/servicecomb-kie/test"
	"github.com/stretchr/testify/assert"
)

func TestGetHistory(t *testing.T) {
	kv, err := kvsvc.Create(context.TODO(), &model.KVDoc{
		Key:    "history",
		Value:  "2s",
		Status: common2.StatusEnabled,
		Labels: map[string]string{
			"app":     "mall",
			"service": "cart",
		},
		Domain:  "default",
		Project: "kv-test",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, kv.ID)
	t.Run("after create kv, should has history", func(t *testing.T) {
		h, err := datasource.GetBroker().GetHistoryDao().GetHistory(context.TODO(), kv.ID, "kv-test", "default")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, h.Total, 1)
	})

}
